package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection
var ctx = context.TODO()

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("timer_db").Collection("timer")
}

func main() {

	app := &cli.App{
		Name:  "timer",
		Usage: "A Simple CLI application to log the time spent on specific tasks",
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "Start work on a task",
				Action: func(c *cli.Context) error {
					str := c.Args().First()
					if str == "" {
						return errors.New("Cannot add an empty task")
					}

					timeLog := &TimeLog{
						ID:        primitive.NewObjectID(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
						StartTime: time.Now(),
						// EndTime : nil,
						Status: "STARTED",
						//intializing empty slices here
						PausedTime: []time.Time{},
						ResumeTime: []time.Time{},
						WorkType:   str,
					}

					return createTimeLog(timeLog)
				},
			},
			{
				Name:    "pause",
				Aliases: []string{"p"},
				Usage:   "pause/resume",
				Action: func(c *cli.Context) error {
					str := c.Args().First()
					if str == "" {
						str = "@"
					}
					return pauseTaskProcessing(str)
				},
			},
			{
				Name:    "resume",
				Aliases: []string{"r"},
				Usage:   "resume",
				Action: func(c *cli.Context) error {
					str := c.Args().First()
					if str == "" {
						str = "@"
					}
					return resumeTaskProcessing(str)
				},
			},
			{
				Name:    "stop",
				Aliases: []string{"s"},
				Usage:   "stop",
				Action: func(c *cli.Context) error {
					str := c.Args().First()
					if str == "" {
						str = "@"
					}
					return stopTaskProcessing(str)
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

type TimeLog struct {
	ID         primitive.ObjectID `bson:"_id"`
	CreatedAt  time.Time          `bson:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt"`
	StartTime  time.Time          `bson:"startTime"`
	EndTime    time.Time          `bson:"endTime"`
	PausedTime []time.Time        `bson:"pausedTime"`
	ResumeTime []time.Time        `bson:"resumeTime"`
	Status     string             `bson:"status"`
	WorkType   string             `bson:"workType"`
}

func createTimeLog(timeLog *TimeLog) error {
	_, err := collection.InsertOne(ctx, timeLog)
	return err
}

func pauseTaskProcessing(workType string) error {
	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"workType", workType}},
				bson.D{{"status", bson.D{{"$ne", "STOPPED"}}}}},
		}}

	update := bson.M{
		"$set": bson.M{
			"status": "PAUSED",
		},
		"$push": bson.M{
			"pausedTime": time.Now(),
		},
	}
	_, err := collection.UpdateMany(ctx, filter, update)
	return err
}

func resumeTaskProcessing(workType string) error {
	filter := bson.D{
		{Key: "$and",
			Value: bson.A{
				bson.D{{Key: "workType", Value: workType}},
				bson.D{{Key: "status", Value: "PAUSED"}}},
		}}

	update := bson.M{
		"$set": bson.M{
			"status": "STARTED",
		},
		"$push": bson.M{
			"resumeTime": time.Now(), // The value you want to add to the array.
		},
	}
	_, err := collection.UpdateMany(ctx, filter, update)
	return err
}

func stopTaskProcessing(workType string) error {
	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"workType", workType}},
				bson.D{{"status", bson.D{{"$ne", "STOPPED"}}}}},
		}}

	update := bson.D{{"$set",
		bson.D{{"status", "STOPPED"}, {"endTime", time.Now()}}}}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}
