package main

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Reviews live in MongoDB (collection `reviews`), one document per review with
// tags embedded as an array — the variable tag payload is why this service left
// Postgres. No cross-document invariant exists here, so no transactions needed.
//
// ponytail: indexes are created idempotently on boot instead of a migration tool.
func ensureIndexes(ctx context.Context, coll *mongo.Collection) error {
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		// GET /api/reviews?library_id=… (public list, newest first)
		{Keys: bson.D{{Key: "library_id", Value: 1}, {Key: "hidden", Value: 1}, {Key: "created_at", Value: -1}}},
		// GET /api/reviews/flagged (admin queue, hidden first)
		{Keys: bson.D{{Key: "hidden", Value: -1}, {Key: "created_at", Value: -1}}},
	})
	return err
}
