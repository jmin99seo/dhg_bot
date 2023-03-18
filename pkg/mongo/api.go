package mongo

import (
	"context"
	"time"

	"github.com/jm199seo/dhg_bot/pkg/loa_api"
	"github.com/jm199seo/dhg_bot/util/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MainCharacters struct {
	ID   primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	Name string             `bson:"name" json:"name"`
}

func (m *Client) MainCharacters(ctx context.Context) ([]MainCharacters, error) {
	var mainCharacters []MainCharacters

	db := m.client.Database(m.Config.DatabaseName)
	coll := db.Collection(m.Config.MainCharactersCollection)

	cur, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err := cur.All(ctx, &mainCharacters); err != nil {
		return nil, err
	}
	return mainCharacters, nil
}

type Character struct {
	ID                primitive.ObjectID    `bson:"_id,omitempty"`
	CharacterInfo     loa_api.CharacterInfo `bson:"character_info"`
	MainCharacterName string                `bson:"main_character_name"`
	CreatedAt         time.Time             `bson:"created_at"`
	UpdatedAt         time.Time             `bson:"updated_at"`
}

func (m *Client) SaveSubCharacters(ctx context.Context, mainCharName string, chars []loa_api.CharacterInfo) error {
	db := m.client.Database(m.Config.DatabaseName)
	coll := db.Collection(m.Config.CharactersCollection)

	var insert []interface{}
	now := time.Now()

	for _, char := range chars {
		c := Character{
			CharacterInfo:     char,
			MainCharacterName: mainCharName,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		insert = append(insert, c)
	}
	_, err := coll.InsertMany(ctx, insert, nil)
	if err != nil {
		logger.Log.Errorf("failed to insert sub characters: %v", err)
		return err
	}
	return nil
}

// 카단서버
func (m *Client) SubCharactersForMainCharacter(ctx context.Context, mainCharName string) ([]Character, error) {
	var characters []Character

	db := m.client.Database(m.Config.DatabaseName)
	coll := db.Collection(m.Config.CharactersCollection)

	cur, err := coll.Find(ctx, bson.M{"main_character_name": mainCharName})
	if err != nil {
		return nil, err
	}
	if err := cur.All(ctx, &characters); err != nil {
		return nil, err
	}
	return characters, nil
}

func (m *Client) UpdateChracter(ctx context.Context, char Character) error {
	db := m.client.Database(m.Config.DatabaseName)
	coll := db.Collection(m.Config.CharactersCollection)

	char.UpdatedAt = time.Now()
	upsertTrue := true

	_, err := coll.UpdateOne(ctx, bson.M{"_id": char.ID}, bson.M{"$set": char}, &options.UpdateOptions{
		Upsert: &upsertTrue,
	})
	if err != nil {
		logger.Log.Errorf("failed to update character: %v", err)
		return err
	}
	return nil
}

func (m *Client) DeleteCharacters(ctx context.Context, chars []Character) error {
	db := m.client.Database(m.Config.DatabaseName)
	coll := db.Collection(m.Config.CharactersCollection)

	var ids []primitive.ObjectID
	for _, char := range chars {
		ids = append(ids, char.ID)
	}
	if _, err := coll.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}}); err != nil {
		logger.Log.Errorf("failed to delete characters: %v", err)
		return err
	}
	return nil
}
