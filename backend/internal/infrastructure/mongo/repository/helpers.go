package repository

import (
	"regexp"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	return page, pageSize
}

func findOptions(page, pageSize int, sort bson.D) *options.FindOptionsBuilder {
	page, pageSize = normalizePage(page, pageSize)
	return options.Find().
		SetSort(sort).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))
}

func keywordRegex(keyword string) bson.Regex {
	return bson.Regex{
		Pattern: regexp.QuoteMeta(keyword),
		Options: "i",
	}
}
