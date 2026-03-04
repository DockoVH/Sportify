package db

import (
	"context"
	"time"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Oglas struct {
	ID bson.ObjectID `bson:"_id,omitempty"`
	Vlasnik string `bson:"vlasnik"`
	Sport string `bson:"sport"`
	Opis string `bson:"opis,omitempty"`
	PotrebnoIgraca int `bson:"potrebno_igraca"`
	Mesto string `bson:"mesto"`
	Koordinate [2]float64 `bson:"koordinate,omitempty"`
	Vreme time.Time `bson:"vreme"`
	Komentari []Komentar `bson:"komentari,omitempty"`
}

type Komentar struct {
	UUID string `bson:"uuid"`
	Vlasnik string `bson:"vlasnik"`
	Sadrzaj string `bson:"sadrzaj"`
	Vreme time.Time `bson:"vreme"`
}

func VratiOglas(mongoClient *mongo.Client, ID bson.ObjectID) (*Oglas, error) {
	oglasi := mongoClient.Database("sportify").Collection("oglasi")

	var oglas Oglas
	if err := oglasi.FindOne(context.Background(), bson.M{ "_id": ID }).Decode(&oglas); err != nil {
		return nil, err
	}

	return &oglas, nil
}

func DodajOglas(mongoClient *mongo.Client, oglas Oglas) (bson.ObjectID, error) {
	oglasi := mongoClient.Database("sportify").Collection("oglasi")

	rezultat, err := oglasi.InsertOne(context.Background(), oglas)
	if err != nil {
		return bson.NilObjectID, err
	}

	return rezultat.InsertedID.(bson.ObjectID), nil
}

func IzmeniOglas(mongoClient *mongo.Client, oglas Oglas) error {
	oglasi := mongoClient.Database("sportify").Collection("oglasi")
	ctx := context.Background()

	filter := bson.M{ "_id": oglas.ID }
	update := bson.M{ "$set": oglas }

	_, err := oglasi.UpdateOne(ctx, filter, update)
	return err
}

func ObrisiOglas(mongoClient *mongo.Client, ID bson.ObjectID) (int64, error) {
	oglasi := mongoClient.Database("sportify").Collection("oglasi")

	rezultat, err := oglasi.DeleteOne(context.Background(), bson.M{ "_id": ID })
	if err != nil {
		return 0, err
	}

	return rezultat.DeletedCount, nil
}

func SviOglasiKorisnika(mongoClient *mongo.Client, username string) ([]Oglas, error) {
	oglasi := mongoClient.Database("sportify").Collection("oglasi")
    ctx := context.Background()

    cursor, err := oglasi.Find(ctx, bson.M{ "vlasnik": username })
    if err != nil {
        return make([]Oglas, 0), err
    }

    var sviOglasi []Oglas
    err = cursor.All(ctx, &sviOglasi)

    return sviOglasi, err
}

func SviOglasi(mongoClient *mongo.Client) ([]Oglas, error) {
	oglasi := mongoClient.Database("sportify").Collection("oglasi")
    ctx := context.Background()

    cursor, err := oglasi.Find(ctx, bson.D{{}})
    if err != nil {
        return make([]Oglas, 0), err
    }

    var sviOglasi []Oglas
    err = cursor.All(ctx, &sviOglasi)

    return sviOglasi, err
}

func VratiKomentar(mongoClient *mongo.Client, oglasID bson.ObjectID, komentarUUID string) (*Komentar, error) {
	oglasi := mongoClient.Database("sportify").Collection("oglasi")

	ctx := context.Background()
	filter := bson.M { "_id": oglasID }
	projekcija := bson.M {
		"komentari": bson.M {
			"$elemMatch": bson.M {
				"uuid": komentarUUID,
			},
		},
	}

	var rezultat struct {
		Komentari []Komentar `bson:"komentari"`
	}

	err := oglasi.FindOne(
		ctx,
		filter,
		options.FindOne().SetProjection(projekcija),
	).Decode(&rezultat)

	if err != nil {
		return nil, err
	}

	if len(rezultat.Komentari) == 0 {
		return nil, errors.New("Ne postoji komentar sa datim UUID-em.")
	}

	return &rezultat.Komentari[0], nil
}

func DodajKomentar(mongoClient *mongo.Client, oglasID bson.ObjectID, komentar Komentar) error {
	oglasi := mongoClient.Database("sportify").Collection("oglasi")

	filter := bson.M { "_id": oglasID }
	update := bson.M {
		"$push": bson.M {
			"komentari": komentar,
		},
	}

	rezultat, err := oglasi.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if rezultat.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func IzmeniKomentar(mongoClient *mongo.Client, oglasID bson.ObjectID, komentar Komentar) error {
	oglasi := mongoClient.Database("sportify").Collection("oglasi")

	filter := bson.M {
		"_id": oglasID,
		"komentari": bson.M {
			"$elemMatch": bson.M {
				"uuid": komentar.UUID,
			},
		},
	}
	update := bson.M {
		"$set": bson.M {
			"komentari.$.sadrzaj": komentar.Sadrzaj,
			"komentari.$.vreme": time.Now(),
		},
	}

	rezultat, err := oglasi.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if rezultat.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func ObrisiKomentar(mongoClient *mongo.Client, oglasID bson.ObjectID, komentarUUID string) error {
	oglasi := mongoClient.Database("sportify").Collection("oglasi")

	filter := bson.M { "_id": oglasID }
	update := bson.M {
		"$pull": bson.M {
			"komentari": bson.M {
				"uuid": komentarUUID,
			},
		},
	}

	rezultat, err := oglasi.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if rezultat.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
