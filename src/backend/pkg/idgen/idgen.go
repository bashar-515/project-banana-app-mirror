package idgen

import (
	"strconv"
	"sync/atomic"

	"github.com/google/uuid"
)

type IdGenerator struct {
	namespace uuid.UUID
	counter   atomic.Int64
}

func NewIdGenerator() *IdGenerator {
	return &IdGenerator{
		namespace: uuid.New(),
	}
}

func NewIdGeneratorWithNamespace(namespace string) (*IdGenerator, error) {
	namespaceParsed, err := uuid.Parse(namespace)
	if err != nil {
		return nil, err
	}

	return &IdGenerator{namespace: namespaceParsed}, nil
}

func (ig *IdGenerator) GenerateId() string {
	return uuid.NewSHA1(ig.namespace, strconv.AppendInt(nil, ig.counter.Add(1)-1, 10)).String()
}
