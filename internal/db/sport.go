package db

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Sport struct {
	Naziv string `bson:"_id"`
	PotrebnoIgraca int `bson:"potrebno_igraca"`
}

func VratiSport(mongoClient *mongo.Client, naziv string) (*Sport, error) {
	sportovi := mongoClient.Database("sportify").Collection("sportovi")

	var sport Sport
	if err := sportovi.FindOne(context.Background(), bson.M { "_id": naziv }).Decode(&sport); err != nil {
		return nil, err
	}

	return &sport, nil
}

func DodajSport(mongoClient *mongo.Client, sport Sport) (string, error) {
	sportovi := mongoClient.Database("sportify").Collection("sportovi")

	rezultat, err := sportovi.InsertOne(context.Background(), sport)
	if err != nil {
		return "", err
	}

	return rezultat.InsertedID.(string), nil
}

func IzmeniSport(mongoClient *mongo.Client, sport Sport) error {
	sportovi := mongoClient.Database("sportify").Collection("sportovi")
	ctx := context.Background()

	filter := bson.M { "_id": sport.Naziv }
	update := bson.M { "$set": sport }

	rezultat, err := sportovi.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if rezultat.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func ObrisiSport(mongoClient *mongo.Client, naziv string) (int64, error) {
	sportovi := mongoClient.Database("sportify").Collection("sportovi")

	rezultat, err := sportovi.DeleteOne(context.Background(), bson.M { "_id": naziv })
	if err != nil {
		return 0, err
	}

	return rezultat.DeletedCount, nil

}

func SviSportovi(mongoClient *mongo.Client) ([]Sport, error) {
	sportovi := mongoClient.Database("sportify").Collection("sportovi")
	ctx := context.Background()

	cursor, err := sportovi.Find(ctx, bson.D {{}})
	if err != nil {
		return make([]Sport, 0), err
	}

	var sviSportovi []Sport
	err = cursor.All(ctx, &sviSportovi)

	return sviSportovi, err
}
