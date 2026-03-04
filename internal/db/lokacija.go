package db

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Lokacija struct {
	Mesto string `bson:"_id"`
}

func VratiLokaciju(mongoClient *mongo.Client, mesto string) (*Lokacija, error) {
	lokacije := mongoClient.Database("sportify").Collection("lokacije")

	var lokacija Lokacija
	if err := lokacije.FindOne(context.Background(), bson.M { "_id": mesto }).Decode(&lokacija); err != nil {
		return nil, err
	}

	return &lokacija, nil
}

func DodajLokaciju(mongoClient *mongo.Client, lokacija Lokacija) (string, error) {
	lokacije := mongoClient.Database("sportify").Collection("lokacije")

	rezultat, err := lokacije.InsertOne(context.Background(), lokacija)
	if err != nil {
		return "", err
	}

	return rezultat.InsertedID.(string), nil
}

func IzmeniLokaciju(mongoClient *mongo.Client, lokacija Lokacija) error {
	lokacije := mongoClient.Database("sportify").Collection("lokacije")
	ctx := context.Background()

	filter := bson.M { "_id": lokacija.Mesto }
	update := bson.M { "$set": lokacija }

	rezultat, err := lokacije.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if rezultat.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func ObrisiLokaciju(mongoClient *mongo.Client, mesto string) (int64, error) {
	lokacije := mongoClient.Database("sportify").Collection("lokacije")

	rezultat, err := lokacije.DeleteOne(context.Background(), bson.M { "_id": mesto })
	if err != nil {
		return 0, err
	}

	return rezultat.DeletedCount, nil

}

func SveLokacije(mongoClient *mongo.Client) (sveLokacije []Lokacija, err error) {
	lokacije := mongoClient.Database("sportify").Collection("lokacije")
	ctx := context.Background()

	cursor, err := lokacije.Find(ctx, bson.D {{}})
	if err != nil {
		return sveLokacije, err
	}

	err = cursor.All(ctx, &sveLokacije)

	return sveLokacije, err
}
