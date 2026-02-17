package db

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Korisnik struct {
	Username string `bson:"_id"`
	LozinkaHes string `bson:"lozinka_hes"`
	SlikaBase64 string `bson:"slika_base64"`
	Ocene []Ocena `bson:"ocene,omitempty"`
}

type Ocena struct {
	UUID string `bson:"uuid"`
	Vlasnik string `bson:"vlasnik"`
	Vrednost uint `bson:"vrednost"`
}

func VratiKorisnika(mongoClient *mongo.Client, username string) (*Korisnik, error){
	korisnici := mongoClient.Database("sportify").Collection("korisnici")

	var korisnik Korisnik
	if err := korisnici.FindOne(context.Background(), bson.M { "_id": username }).Decode(&korisnik); err != nil {
		return nil, err
	}

	return &korisnik, nil
}

func DodajKorisnika(mongoClient *mongo.Client, korisnik Korisnik) (string, error) {
	korisnici := mongoClient.Database("sportify").Collection("korisnici")

	rezultat, err := korisnici.InsertOne(context.Background(), korisnik)
	if err != nil {
		return "", err
	}

	return rezultat.InsertedID.(string), nil
}

func IzmeniKorisnika(mongoClient *mongo.Client, korisnik Korisnik) error {
	korisnici := mongoClient.Database("sportify").Collection("korisnici")
	ctx := context.Background()
	var stariPodaci Korisnik

	filter := bson.M{ "_id": korisnik.Username }
	update := bson.M{ "$set": korisnik }

	return korisnici.FindOneAndUpdate(ctx, filter, update).Decode(&stariPodaci)
}

func ObrisiKorisnika(mongoClient *mongo.Client, username string) (int64, error) {
	korisnici := mongoClient.Database("sportify").Collection("korisnici")

	rezultat, err := korisnici.DeleteOne(context.Background(), bson.M{ "_id": username })
	if err != nil {
		return 0, err
	}

	return rezultat.DeletedCount, nil
}

func SviKorisnici(mongoClient *mongo.Client) ([]Korisnik, error) {
	korisnici := mongoClient.Database("sportify").Collection("korisnici")
	ctx := context.Background()

	cursor, err := korisnici.Find(ctx, bson.D{{}})
	if err != nil {
		return make([]Korisnik, 0), err
	}

	var sviKorisnici []Korisnik
	err = cursor.All(ctx, &sviKorisnici)

	return sviKorisnici, err
}

func VratiOcenuKorisnika(mongoClient *mongo.Client, username string, ocenaUUID string) (*Ocena, error) {
	korisnici := mongoClient.Database("sportify").Collection("korisnici")

    ctx := context.Background()
    filter := bson.M { "_id": username }
    projekcija := bson.M {
        "ocene": bson.M {
            "$elemMatch": bson.M {
                "uuid": ocenaUUID,
            },
        },
    }

    var rezultat struct {
        Ocene []Ocena `bson:"ocene"`
    }

    err := korisnici.FindOne(
        ctx,
        filter,
        options.FindOne().SetProjection(projekcija),
    ).Decode(&rezultat)

    if err != nil {
        return nil, err
    }

    if len(rezultat.Ocene) == 0 {
        return nil, errors.New("Ne postoji ocena sa datim UUID-em.")
    }

    return &rezultat.Ocene[0], nil
}

func DodajOcenuKorisnika(mongoClient *mongo.Client, username string, ocena Ocena) error  {
	korisnici := mongoClient.Database("sportify").Collection("korisnici")

    ctx := context.Background()
    filter := bson.M { "_id": username }
	update := bson.M {
		"$push": bson.M {
			"ocene": ocena,
		},
	}

	rezultat, err := korisnici.UpdateOne(ctx, filter, update)
    if err != nil {
        return err
    }

    if rezultat.MatchedCount == 0 {
        return mongo.ErrNoDocuments
    }

    return nil
}

func IzmeniOcenuKorisnika(mongoClient *mongo.Client, username string, ocena Ocena) error {
	korisnici := mongoClient.Database("sportify").Collection("korisnici")

    ctx := context.Background()
    filter := bson.M {
		"_id": username,
		"ocene": bson.M {
			"$elemMatch": bson.M {
				"uuid": ocena.UUID,
			},
		},
	}
    update := bson.M {
        "$set": bson.M {
			"ocene.$.vrednost": ocena.Vrednost,
        },
    }

	rezultat, err := korisnici.UpdateOne(ctx, filter, update)
    if err != nil {
        return nil
    }

    if rezultat.MatchedCount == 0 {
        return mongo.ErrNoDocuments
    }

    return nil
}

func ObrisiOcenuKorisnika(mongoClient *mongo.Client, username string, ocenaUUID string) error {
	korisnici := mongoClient.Database("sportify").Collection("korisnici")

	ctx := context.Background()
	filter := bson.M { "_id": username }
	update := bson.M {
		"$pull": bson.M {
			"ocene": bson.M {
				"uuid": ocenaUUID,
			},
		},
	}

	rezultat, err := korisnici.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if rezultat.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
