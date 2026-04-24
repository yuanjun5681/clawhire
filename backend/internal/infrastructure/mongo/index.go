package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type indexSpec struct {
	collection string
	name       string
	model      mongo.IndexModel
}

type legacyIndexSpec struct {
	collection string
	name       string
}

func idx(coll, name string, keys bson.D) indexSpec {
	return indexSpec{
		collection: coll,
		name:       name,
		model:      mongo.IndexModel{Keys: keys, Options: options.Index().SetName(name)},
	}
}

func uniqueIdx(coll, name string, keys bson.D) indexSpec {
	return indexSpec{
		collection: coll,
		name:       name,
		model:      mongo.IndexModel{Keys: keys, Options: options.Index().SetName(name).SetUnique(true)},
	}
}

func sparseUniqueIdx(coll, name string, keys bson.D) indexSpec {
	return indexSpec{
		collection: coll,
		name:       name,
		model:      mongo.IndexModel{Keys: keys, Options: options.Index().SetName(name).SetUnique(true).SetSparse(true)},
	}
}

func partialUniqueIdx(coll, name string, keys bson.D, filter interface{}) indexSpec {
	return indexSpec{
		collection: coll,
		name:       name,
		model: mongo.IndexModel{
			Keys:    keys,
			Options: options.Index().SetName(name).SetUnique(true).SetPartialFilterExpression(filter),
		},
	}
}

func indexSpecs() []indexSpec {
	return []indexSpec{
		// accounts
		uniqueIdx(CollAccounts, "uk_accountId", bson.D{{Key: "accountId", Value: 1}}),
		idx(CollAccounts, "ix_type_createdAt", bson.D{{Key: "type", Value: 1}, {Key: "createdAt", Value: -1}}),
		sparseUniqueIdx(CollAccounts, "uk_nodeId", bson.D{{Key: "nodeId", Value: 1}}),
		idx(CollAccounts, "ix_ownerAccountId_createdAt", bson.D{{Key: "ownerAccountId", Value: 1}, {Key: "createdAt", Value: -1}}),

		// tasks
		uniqueIdx(CollTasks, "uk_taskId", bson.D{{Key: "taskId", Value: 1}}),
		idx(CollTasks, "ix_status_createdAt", bson.D{{Key: "status", Value: 1}, {Key: "createdAt", Value: -1}}),
		idx(CollTasks, "ix_requester_createdAt", bson.D{{Key: "requester.id", Value: 1}, {Key: "createdAt", Value: -1}}),
		idx(CollTasks, "ix_executor_createdAt", bson.D{{Key: "assignedExecutor.id", Value: 1}, {Key: "createdAt", Value: -1}}),
		idx(CollTasks, "ix_category_status_createdAt", bson.D{{Key: "category", Value: 1}, {Key: "status", Value: 1}, {Key: "createdAt", Value: -1}}),
		idx(CollTasks, "ix_lastActivityAt", bson.D{{Key: "lastActivityAt", Value: -1}}),

		// bids
		uniqueIdx(CollBids, "uk_bidId", bson.D{{Key: "bidId", Value: 1}}),
		idx(CollBids, "ix_taskId_createdAt", bson.D{{Key: "taskId", Value: 1}, {Key: "createdAt", Value: -1}}),
		idx(CollBids, "ix_executor_createdAt", bson.D{{Key: "executor.id", Value: 1}, {Key: "createdAt", Value: -1}}),

		// contracts
		uniqueIdx(CollContracts, "uk_contractId", bson.D{{Key: "contractId", Value: 1}}),
		partialUniqueIdx(CollContracts, "uk_active_contract_per_task", bson.D{{Key: "taskId", Value: 1}}, bson.M{"status": "active"}),
		idx(CollContracts, "ix_executor_createdAt", bson.D{{Key: "executor.id", Value: 1}, {Key: "createdAt", Value: -1}}),

		// progress_reports
		uniqueIdx(CollProgress, "uk_progressId", bson.D{{Key: "progressId", Value: 1}}),
		idx(CollProgress, "ix_taskId_reportedAt", bson.D{{Key: "taskId", Value: 1}, {Key: "reportedAt", Value: -1}}),

		// milestones
		uniqueIdx(CollMilestones, "uk_milestoneId", bson.D{{Key: "milestoneId", Value: 1}}),
		idx(CollMilestones, "ix_taskId_reportedAt", bson.D{{Key: "taskId", Value: 1}, {Key: "reportedAt", Value: -1}}),

		// submissions
		uniqueIdx(CollSubmissions, "uk_submissionId", bson.D{{Key: "submissionId", Value: 1}}),
		idx(CollSubmissions, "ix_taskId_submittedAt", bson.D{{Key: "taskId", Value: 1}, {Key: "submittedAt", Value: -1}}),

		// reviews
		uniqueIdx(CollReviews, "uk_reviewId", bson.D{{Key: "reviewId", Value: 1}}),
		idx(CollReviews, "ix_taskId_reviewedAt", bson.D{{Key: "taskId", Value: 1}, {Key: "reviewedAt", Value: -1}}),
		idx(CollReviews, "ix_submissionId", bson.D{{Key: "submissionId", Value: 1}}),

		// settlements
		uniqueIdx(CollSettlements, "uk_settlementId", bson.D{{Key: "settlementId", Value: 1}}),
		idx(CollSettlements, "ix_taskId_recordedAt", bson.D{{Key: "taskId", Value: 1}, {Key: "recordedAt", Value: -1}}),
		idx(CollSettlements, "ix_payee_recordedAt", bson.D{{Key: "payee.id", Value: 1}, {Key: "recordedAt", Value: -1}}),

		// raw_events
		uniqueIdx(CollRawEvents, "uk_eventKey", bson.D{{Key: "eventKey", Value: 1}}),
		idx(CollRawEvents, "ix_messageType_receivedAt", bson.D{{Key: "messageType", Value: 1}, {Key: "receivedAt", Value: -1}}),

		// domain_events
		uniqueIdx(CollDomainEvents, "uk_eventId", bson.D{{Key: "eventId", Value: 1}}),
		idx(CollDomainEvents, "ix_aggregate_createdAt", bson.D{{Key: "aggregateType", Value: 1}, {Key: "aggregateId", Value: 1}, {Key: "createdAt", Value: -1}}),
	}
}

func legacyIndexSpecs() []legacyIndexSpec {
	return []legacyIndexSpec{
		// `contracts.taskId` 旧版曾使用普通索引 `ix_taskId`。
		// 现在该约束已被“每个 task 最多一个 active contract”的部分唯一索引替代，
		// 两者 key pattern 相同，Mongo 无法并存，因此启动时需要先清理遗留索引。
		{collection: CollContracts, name: "ix_taskId"},
	}
}

func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	if err := dropLegacyIndexes(ctx, db); err != nil {
		return err
	}
	for _, s := range indexSpecs() {
		if _, err := db.Collection(s.collection).Indexes().CreateOne(ctx, s.model); err != nil {
			return fmt.Errorf("create index %s on %s: %w", s.name, s.collection, err)
		}
	}
	return nil
}

func dropLegacyIndexes(ctx context.Context, db *mongo.Database) error {
	for _, legacy := range legacyIndexSpecs() {
		coll := db.Collection(legacy.collection)
		specs, err := coll.Indexes().ListSpecifications(ctx)
		if err != nil {
			return fmt.Errorf("list indexes on %s: %w", legacy.collection, err)
		}

		found := false
		for _, spec := range specs {
			if spec.Name == legacy.name {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		if err := coll.Indexes().DropOne(ctx, legacy.name); err != nil {
			return fmt.Errorf("drop legacy index %s on %s: %w", legacy.name, legacy.collection, err)
		}
	}
	return nil
}
