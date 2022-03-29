// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-files-api/mongo"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	lock "github.com/square/mongo-lock"
	"sync"
)

// Ensure, that MongoCollectionMock does implement mongo.MongoCollection.
// If this is not the case, regenerate this file with moq.
var _ mongo.MongoCollection = &MongoCollectionMock{}

// MongoCollectionMock is a mock implementation of mongo.MongoCollection.
//
// 	func TestSomethingThatUsesMongoCollection(t *testing.T) {
//
// 		// make and configure a mocked mongo.MongoCollection
// 		mockedMongoCollection := &MongoCollectionMock{
// 			AggregateFunc: func(ctx context.Context, pipeline interface{}, results interface{}) error {
// 				panic("mock out the Aggregate method")
// 			},
// 			CountFunc: func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
// 				panic("mock out the Count method")
// 			},
// 			DeleteFunc: func(ctx context.Context, selector interface{}) (*mongodriver.CollectionDeleteResult, error) {
// 				panic("mock out the Delete method")
// 			},
// 			DeleteByIdFunc: func(ctx context.Context, id interface{}) (*mongodriver.CollectionDeleteResult, error) {
// 				panic("mock out the DeleteById method")
// 			},
// 			DeleteManyFunc: func(ctx context.Context, selector interface{}) (*mongodriver.CollectionDeleteResult, error) {
// 				panic("mock out the DeleteMany method")
// 			},
// 			DistinctFunc: func(ctx context.Context, fieldName string, filter interface{}) ([]interface{}, error) {
// 				panic("mock out the Distinct method")
// 			},
// 			FindFunc: func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
// 				panic("mock out the Find method")
// 			},
// 			FindOneFunc: func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
// 				panic("mock out the FindOne method")
// 			},
// 			InsertFunc: func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
// 				panic("mock out the Insert method")
// 			},
// 			InsertManyFunc: func(ctx context.Context, documents []interface{}) (*mongodriver.CollectionInsertManyResult, error) {
// 				panic("mock out the InsertMany method")
// 			},
// 			MustFunc: func() *mongodriver.Must {
// 				panic("mock out the Must method")
// 			},
// 			NewLockClientFunc: func() *lock.Client {
// 				panic("mock out the NewLockClient method")
// 			},
// 			UpdateFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
// 				panic("mock out the Update method")
// 			},
// 			UpdateByIdFunc: func(ctx context.Context, id interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
// 				panic("mock out the UpdateById method")
// 			},
// 			UpdateManyFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
// 				panic("mock out the UpdateMany method")
// 			},
// 			UpsertFunc: func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
// 				panic("mock out the Upsert method")
// 			},
// 			UpsertByIdFunc: func(ctx context.Context, id interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
// 				panic("mock out the UpsertById method")
// 			},
// 		}
//
// 		// use mockedMongoCollection in code that requires mongo.MongoCollection
// 		// and then make assertions.
//
// 	}
type MongoCollectionMock struct {
	// AggregateFunc mocks the Aggregate method.
	AggregateFunc func(ctx context.Context, pipeline interface{}, results interface{}) error

	// CountFunc mocks the Count method.
	CountFunc func(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error)

	// DeleteFunc mocks the Delete method.
	DeleteFunc func(ctx context.Context, selector interface{}) (*mongodriver.CollectionDeleteResult, error)

	// DeleteByIdFunc mocks the DeleteById method.
	DeleteByIdFunc func(ctx context.Context, id interface{}) (*mongodriver.CollectionDeleteResult, error)

	// DeleteManyFunc mocks the DeleteMany method.
	DeleteManyFunc func(ctx context.Context, selector interface{}) (*mongodriver.CollectionDeleteResult, error)

	// DistinctFunc mocks the Distinct method.
	DistinctFunc func(ctx context.Context, fieldName string, filter interface{}) ([]interface{}, error)

	// FindFunc mocks the Find method.
	FindFunc func(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error)

	// FindOneFunc mocks the FindOne method.
	FindOneFunc func(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error

	// InsertFunc mocks the Insert method.
	InsertFunc func(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error)

	// InsertManyFunc mocks the InsertMany method.
	InsertManyFunc func(ctx context.Context, documents []interface{}) (*mongodriver.CollectionInsertManyResult, error)

	// MustFunc mocks the Must method.
	MustFunc func() *mongodriver.Must

	// NewLockClientFunc mocks the NewLockClient method.
	NewLockClientFunc func() *lock.Client

	// UpdateFunc mocks the Update method.
	UpdateFunc func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error)

	// UpdateByIdFunc mocks the UpdateById method.
	UpdateByIdFunc func(ctx context.Context, id interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error)

	// UpdateManyFunc mocks the UpdateMany method.
	UpdateManyFunc func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error)

	// UpsertFunc mocks the Upsert method.
	UpsertFunc func(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error)

	// UpsertByIdFunc mocks the UpsertById method.
	UpsertByIdFunc func(ctx context.Context, id interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error)

	// calls tracks calls to the methods.
	calls struct {
		// Aggregate holds details about calls to the Aggregate method.
		Aggregate []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Pipeline is the pipeline argument value.
			Pipeline interface{}
			// Results is the results argument value.
			Results interface{}
		}
		// Count holds details about calls to the Count method.
		Count []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Filter is the filter argument value.
			Filter interface{}
			// Opts is the opts argument value.
			Opts []mongodriver.FindOption
		}
		// Delete holds details about calls to the Delete method.
		Delete []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Selector is the selector argument value.
			Selector interface{}
		}
		// DeleteById holds details about calls to the DeleteById method.
		DeleteById []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ID is the id argument value.
			ID interface{}
		}
		// DeleteMany holds details about calls to the DeleteMany method.
		DeleteMany []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Selector is the selector argument value.
			Selector interface{}
		}
		// Distinct holds details about calls to the Distinct method.
		Distinct []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// FieldName is the fieldName argument value.
			FieldName string
			// Filter is the filter argument value.
			Filter interface{}
		}
		// Find holds details about calls to the Find method.
		Find []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Filter is the filter argument value.
			Filter interface{}
			// Results is the results argument value.
			Results interface{}
			// Opts is the opts argument value.
			Opts []mongodriver.FindOption
		}
		// FindOne holds details about calls to the FindOne method.
		FindOne []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Filter is the filter argument value.
			Filter interface{}
			// Result is the result argument value.
			Result interface{}
			// Opts is the opts argument value.
			Opts []mongodriver.FindOption
		}
		// Insert holds details about calls to the Insert method.
		Insert []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Document is the document argument value.
			Document interface{}
		}
		// InsertMany holds details about calls to the InsertMany method.
		InsertMany []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Documents is the documents argument value.
			Documents []interface{}
		}
		// Must holds details about calls to the Must method.
		Must []struct {
		}
		// NewLockClient holds details about calls to the NewLockClient method.
		NewLockClient []struct {
		}
		// Update holds details about calls to the Update method.
		Update []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Selector is the selector argument value.
			Selector interface{}
			// Update is the update argument value.
			Update interface{}
		}
		// UpdateById holds details about calls to the UpdateById method.
		UpdateById []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ID is the id argument value.
			ID interface{}
			// Update is the update argument value.
			Update interface{}
		}
		// UpdateMany holds details about calls to the UpdateMany method.
		UpdateMany []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Selector is the selector argument value.
			Selector interface{}
			// Update is the update argument value.
			Update interface{}
		}
		// Upsert holds details about calls to the Upsert method.
		Upsert []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Selector is the selector argument value.
			Selector interface{}
			// Update is the update argument value.
			Update interface{}
		}
		// UpsertById holds details about calls to the UpsertById method.
		UpsertById []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ID is the id argument value.
			ID interface{}
			// Update is the update argument value.
			Update interface{}
		}
	}
	lockAggregate     sync.RWMutex
	lockCount         sync.RWMutex
	lockDelete        sync.RWMutex
	lockDeleteById    sync.RWMutex
	lockDeleteMany    sync.RWMutex
	lockDistinct      sync.RWMutex
	lockFind          sync.RWMutex
	lockFindOne       sync.RWMutex
	lockInsert        sync.RWMutex
	lockInsertMany    sync.RWMutex
	lockMust          sync.RWMutex
	lockNewLockClient sync.RWMutex
	lockUpdate        sync.RWMutex
	lockUpdateById    sync.RWMutex
	lockUpdateMany    sync.RWMutex
	lockUpsert        sync.RWMutex
	lockUpsertById    sync.RWMutex
}

// Aggregate calls AggregateFunc.
func (mock *MongoCollectionMock) Aggregate(ctx context.Context, pipeline interface{}, results interface{}) error {
	if mock.AggregateFunc == nil {
		panic("MongoCollectionMock.AggregateFunc: method is nil but MongoCollection.Aggregate was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Pipeline interface{}
		Results  interface{}
	}{
		Ctx:      ctx,
		Pipeline: pipeline,
		Results:  results,
	}
	mock.lockAggregate.Lock()
	mock.calls.Aggregate = append(mock.calls.Aggregate, callInfo)
	mock.lockAggregate.Unlock()
	return mock.AggregateFunc(ctx, pipeline, results)
}

// AggregateCalls gets all the calls that were made to Aggregate.
// Check the length with:
//     len(mockedMongoCollection.AggregateCalls())
func (mock *MongoCollectionMock) AggregateCalls() []struct {
	Ctx      context.Context
	Pipeline interface{}
	Results  interface{}
} {
	var calls []struct {
		Ctx      context.Context
		Pipeline interface{}
		Results  interface{}
	}
	mock.lockAggregate.RLock()
	calls = mock.calls.Aggregate
	mock.lockAggregate.RUnlock()
	return calls
}

// Count calls CountFunc.
func (mock *MongoCollectionMock) Count(ctx context.Context, filter interface{}, opts ...mongodriver.FindOption) (int, error) {
	if mock.CountFunc == nil {
		panic("MongoCollectionMock.CountFunc: method is nil but MongoCollection.Count was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Filter interface{}
		Opts   []mongodriver.FindOption
	}{
		Ctx:    ctx,
		Filter: filter,
		Opts:   opts,
	}
	mock.lockCount.Lock()
	mock.calls.Count = append(mock.calls.Count, callInfo)
	mock.lockCount.Unlock()
	return mock.CountFunc(ctx, filter, opts...)
}

// CountCalls gets all the calls that were made to Count.
// Check the length with:
//     len(mockedMongoCollection.CountCalls())
func (mock *MongoCollectionMock) CountCalls() []struct {
	Ctx    context.Context
	Filter interface{}
	Opts   []mongodriver.FindOption
} {
	var calls []struct {
		Ctx    context.Context
		Filter interface{}
		Opts   []mongodriver.FindOption
	}
	mock.lockCount.RLock()
	calls = mock.calls.Count
	mock.lockCount.RUnlock()
	return calls
}

// Delete calls DeleteFunc.
func (mock *MongoCollectionMock) Delete(ctx context.Context, selector interface{}) (*mongodriver.CollectionDeleteResult, error) {
	if mock.DeleteFunc == nil {
		panic("MongoCollectionMock.DeleteFunc: method is nil but MongoCollection.Delete was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Selector interface{}
	}{
		Ctx:      ctx,
		Selector: selector,
	}
	mock.lockDelete.Lock()
	mock.calls.Delete = append(mock.calls.Delete, callInfo)
	mock.lockDelete.Unlock()
	return mock.DeleteFunc(ctx, selector)
}

// DeleteCalls gets all the calls that were made to Delete.
// Check the length with:
//     len(mockedMongoCollection.DeleteCalls())
func (mock *MongoCollectionMock) DeleteCalls() []struct {
	Ctx      context.Context
	Selector interface{}
} {
	var calls []struct {
		Ctx      context.Context
		Selector interface{}
	}
	mock.lockDelete.RLock()
	calls = mock.calls.Delete
	mock.lockDelete.RUnlock()
	return calls
}

// DeleteById calls DeleteByIdFunc.
func (mock *MongoCollectionMock) DeleteById(ctx context.Context, id interface{}) (*mongodriver.CollectionDeleteResult, error) {
	if mock.DeleteByIdFunc == nil {
		panic("MongoCollectionMock.DeleteByIdFunc: method is nil but MongoCollection.DeleteById was just called")
	}
	callInfo := struct {
		Ctx context.Context
		ID  interface{}
	}{
		Ctx: ctx,
		ID:  id,
	}
	mock.lockDeleteById.Lock()
	mock.calls.DeleteById = append(mock.calls.DeleteById, callInfo)
	mock.lockDeleteById.Unlock()
	return mock.DeleteByIdFunc(ctx, id)
}

// DeleteByIdCalls gets all the calls that were made to DeleteById.
// Check the length with:
//     len(mockedMongoCollection.DeleteByIdCalls())
func (mock *MongoCollectionMock) DeleteByIdCalls() []struct {
	Ctx context.Context
	ID  interface{}
} {
	var calls []struct {
		Ctx context.Context
		ID  interface{}
	}
	mock.lockDeleteById.RLock()
	calls = mock.calls.DeleteById
	mock.lockDeleteById.RUnlock()
	return calls
}

// DeleteMany calls DeleteManyFunc.
func (mock *MongoCollectionMock) DeleteMany(ctx context.Context, selector interface{}) (*mongodriver.CollectionDeleteResult, error) {
	if mock.DeleteManyFunc == nil {
		panic("MongoCollectionMock.DeleteManyFunc: method is nil but MongoCollection.DeleteMany was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Selector interface{}
	}{
		Ctx:      ctx,
		Selector: selector,
	}
	mock.lockDeleteMany.Lock()
	mock.calls.DeleteMany = append(mock.calls.DeleteMany, callInfo)
	mock.lockDeleteMany.Unlock()
	return mock.DeleteManyFunc(ctx, selector)
}

// DeleteManyCalls gets all the calls that were made to DeleteMany.
// Check the length with:
//     len(mockedMongoCollection.DeleteManyCalls())
func (mock *MongoCollectionMock) DeleteManyCalls() []struct {
	Ctx      context.Context
	Selector interface{}
} {
	var calls []struct {
		Ctx      context.Context
		Selector interface{}
	}
	mock.lockDeleteMany.RLock()
	calls = mock.calls.DeleteMany
	mock.lockDeleteMany.RUnlock()
	return calls
}

// Distinct calls DistinctFunc.
func (mock *MongoCollectionMock) Distinct(ctx context.Context, fieldName string, filter interface{}) ([]interface{}, error) {
	if mock.DistinctFunc == nil {
		panic("MongoCollectionMock.DistinctFunc: method is nil but MongoCollection.Distinct was just called")
	}
	callInfo := struct {
		Ctx       context.Context
		FieldName string
		Filter    interface{}
	}{
		Ctx:       ctx,
		FieldName: fieldName,
		Filter:    filter,
	}
	mock.lockDistinct.Lock()
	mock.calls.Distinct = append(mock.calls.Distinct, callInfo)
	mock.lockDistinct.Unlock()
	return mock.DistinctFunc(ctx, fieldName, filter)
}

// DistinctCalls gets all the calls that were made to Distinct.
// Check the length with:
//     len(mockedMongoCollection.DistinctCalls())
func (mock *MongoCollectionMock) DistinctCalls() []struct {
	Ctx       context.Context
	FieldName string
	Filter    interface{}
} {
	var calls []struct {
		Ctx       context.Context
		FieldName string
		Filter    interface{}
	}
	mock.lockDistinct.RLock()
	calls = mock.calls.Distinct
	mock.lockDistinct.RUnlock()
	return calls
}

// Find calls FindFunc.
func (mock *MongoCollectionMock) Find(ctx context.Context, filter interface{}, results interface{}, opts ...mongodriver.FindOption) (int, error) {
	if mock.FindFunc == nil {
		panic("MongoCollectionMock.FindFunc: method is nil but MongoCollection.Find was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		Filter  interface{}
		Results interface{}
		Opts    []mongodriver.FindOption
	}{
		Ctx:     ctx,
		Filter:  filter,
		Results: results,
		Opts:    opts,
	}
	mock.lockFind.Lock()
	mock.calls.Find = append(mock.calls.Find, callInfo)
	mock.lockFind.Unlock()
	return mock.FindFunc(ctx, filter, results, opts...)
}

// FindCalls gets all the calls that were made to Find.
// Check the length with:
//     len(mockedMongoCollection.FindCalls())
func (mock *MongoCollectionMock) FindCalls() []struct {
	Ctx     context.Context
	Filter  interface{}
	Results interface{}
	Opts    []mongodriver.FindOption
} {
	var calls []struct {
		Ctx     context.Context
		Filter  interface{}
		Results interface{}
		Opts    []mongodriver.FindOption
	}
	mock.lockFind.RLock()
	calls = mock.calls.Find
	mock.lockFind.RUnlock()
	return calls
}

// FindOne calls FindOneFunc.
func (mock *MongoCollectionMock) FindOne(ctx context.Context, filter interface{}, result interface{}, opts ...mongodriver.FindOption) error {
	if mock.FindOneFunc == nil {
		panic("MongoCollectionMock.FindOneFunc: method is nil but MongoCollection.FindOne was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Filter interface{}
		Result interface{}
		Opts   []mongodriver.FindOption
	}{
		Ctx:    ctx,
		Filter: filter,
		Result: result,
		Opts:   opts,
	}
	mock.lockFindOne.Lock()
	mock.calls.FindOne = append(mock.calls.FindOne, callInfo)
	mock.lockFindOne.Unlock()
	return mock.FindOneFunc(ctx, filter, result, opts...)
}

// FindOneCalls gets all the calls that were made to FindOne.
// Check the length with:
//     len(mockedMongoCollection.FindOneCalls())
func (mock *MongoCollectionMock) FindOneCalls() []struct {
	Ctx    context.Context
	Filter interface{}
	Result interface{}
	Opts   []mongodriver.FindOption
} {
	var calls []struct {
		Ctx    context.Context
		Filter interface{}
		Result interface{}
		Opts   []mongodriver.FindOption
	}
	mock.lockFindOne.RLock()
	calls = mock.calls.FindOne
	mock.lockFindOne.RUnlock()
	return calls
}

// Insert calls InsertFunc.
func (mock *MongoCollectionMock) Insert(ctx context.Context, document interface{}) (*mongodriver.CollectionInsertResult, error) {
	if mock.InsertFunc == nil {
		panic("MongoCollectionMock.InsertFunc: method is nil but MongoCollection.Insert was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Document interface{}
	}{
		Ctx:      ctx,
		Document: document,
	}
	mock.lockInsert.Lock()
	mock.calls.Insert = append(mock.calls.Insert, callInfo)
	mock.lockInsert.Unlock()
	return mock.InsertFunc(ctx, document)
}

// InsertCalls gets all the calls that were made to Insert.
// Check the length with:
//     len(mockedMongoCollection.InsertCalls())
func (mock *MongoCollectionMock) InsertCalls() []struct {
	Ctx      context.Context
	Document interface{}
} {
	var calls []struct {
		Ctx      context.Context
		Document interface{}
	}
	mock.lockInsert.RLock()
	calls = mock.calls.Insert
	mock.lockInsert.RUnlock()
	return calls
}

// InsertMany calls InsertManyFunc.
func (mock *MongoCollectionMock) InsertMany(ctx context.Context, documents []interface{}) (*mongodriver.CollectionInsertManyResult, error) {
	if mock.InsertManyFunc == nil {
		panic("MongoCollectionMock.InsertManyFunc: method is nil but MongoCollection.InsertMany was just called")
	}
	callInfo := struct {
		Ctx       context.Context
		Documents []interface{}
	}{
		Ctx:       ctx,
		Documents: documents,
	}
	mock.lockInsertMany.Lock()
	mock.calls.InsertMany = append(mock.calls.InsertMany, callInfo)
	mock.lockInsertMany.Unlock()
	return mock.InsertManyFunc(ctx, documents)
}

// InsertManyCalls gets all the calls that were made to InsertMany.
// Check the length with:
//     len(mockedMongoCollection.InsertManyCalls())
func (mock *MongoCollectionMock) InsertManyCalls() []struct {
	Ctx       context.Context
	Documents []interface{}
} {
	var calls []struct {
		Ctx       context.Context
		Documents []interface{}
	}
	mock.lockInsertMany.RLock()
	calls = mock.calls.InsertMany
	mock.lockInsertMany.RUnlock()
	return calls
}

// Must calls MustFunc.
func (mock *MongoCollectionMock) Must() *mongodriver.Must {
	if mock.MustFunc == nil {
		panic("MongoCollectionMock.MustFunc: method is nil but MongoCollection.Must was just called")
	}
	callInfo := struct {
	}{}
	mock.lockMust.Lock()
	mock.calls.Must = append(mock.calls.Must, callInfo)
	mock.lockMust.Unlock()
	return mock.MustFunc()
}

// MustCalls gets all the calls that were made to Must.
// Check the length with:
//     len(mockedMongoCollection.MustCalls())
func (mock *MongoCollectionMock) MustCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockMust.RLock()
	calls = mock.calls.Must
	mock.lockMust.RUnlock()
	return calls
}

// NewLockClient calls NewLockClientFunc.
func (mock *MongoCollectionMock) NewLockClient() *lock.Client {
	if mock.NewLockClientFunc == nil {
		panic("MongoCollectionMock.NewLockClientFunc: method is nil but MongoCollection.NewLockClient was just called")
	}
	callInfo := struct {
	}{}
	mock.lockNewLockClient.Lock()
	mock.calls.NewLockClient = append(mock.calls.NewLockClient, callInfo)
	mock.lockNewLockClient.Unlock()
	return mock.NewLockClientFunc()
}

// NewLockClientCalls gets all the calls that were made to NewLockClient.
// Check the length with:
//     len(mockedMongoCollection.NewLockClientCalls())
func (mock *MongoCollectionMock) NewLockClientCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockNewLockClient.RLock()
	calls = mock.calls.NewLockClient
	mock.lockNewLockClient.RUnlock()
	return calls
}

// Update calls UpdateFunc.
func (mock *MongoCollectionMock) Update(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
	if mock.UpdateFunc == nil {
		panic("MongoCollectionMock.UpdateFunc: method is nil but MongoCollection.Update was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Selector interface{}
		Update   interface{}
	}{
		Ctx:      ctx,
		Selector: selector,
		Update:   update,
	}
	mock.lockUpdate.Lock()
	mock.calls.Update = append(mock.calls.Update, callInfo)
	mock.lockUpdate.Unlock()
	return mock.UpdateFunc(ctx, selector, update)
}

// UpdateCalls gets all the calls that were made to Update.
// Check the length with:
//     len(mockedMongoCollection.UpdateCalls())
func (mock *MongoCollectionMock) UpdateCalls() []struct {
	Ctx      context.Context
	Selector interface{}
	Update   interface{}
} {
	var calls []struct {
		Ctx      context.Context
		Selector interface{}
		Update   interface{}
	}
	mock.lockUpdate.RLock()
	calls = mock.calls.Update
	mock.lockUpdate.RUnlock()
	return calls
}

// UpdateById calls UpdateByIdFunc.
func (mock *MongoCollectionMock) UpdateById(ctx context.Context, id interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
	if mock.UpdateByIdFunc == nil {
		panic("MongoCollectionMock.UpdateByIdFunc: method is nil but MongoCollection.UpdateById was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		ID     interface{}
		Update interface{}
	}{
		Ctx:    ctx,
		ID:     id,
		Update: update,
	}
	mock.lockUpdateById.Lock()
	mock.calls.UpdateById = append(mock.calls.UpdateById, callInfo)
	mock.lockUpdateById.Unlock()
	return mock.UpdateByIdFunc(ctx, id, update)
}

// UpdateByIdCalls gets all the calls that were made to UpdateById.
// Check the length with:
//     len(mockedMongoCollection.UpdateByIdCalls())
func (mock *MongoCollectionMock) UpdateByIdCalls() []struct {
	Ctx    context.Context
	ID     interface{}
	Update interface{}
} {
	var calls []struct {
		Ctx    context.Context
		ID     interface{}
		Update interface{}
	}
	mock.lockUpdateById.RLock()
	calls = mock.calls.UpdateById
	mock.lockUpdateById.RUnlock()
	return calls
}

// UpdateMany calls UpdateManyFunc.
func (mock *MongoCollectionMock) UpdateMany(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
	if mock.UpdateManyFunc == nil {
		panic("MongoCollectionMock.UpdateManyFunc: method is nil but MongoCollection.UpdateMany was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Selector interface{}
		Update   interface{}
	}{
		Ctx:      ctx,
		Selector: selector,
		Update:   update,
	}
	mock.lockUpdateMany.Lock()
	mock.calls.UpdateMany = append(mock.calls.UpdateMany, callInfo)
	mock.lockUpdateMany.Unlock()
	return mock.UpdateManyFunc(ctx, selector, update)
}

// UpdateManyCalls gets all the calls that were made to UpdateMany.
// Check the length with:
//     len(mockedMongoCollection.UpdateManyCalls())
func (mock *MongoCollectionMock) UpdateManyCalls() []struct {
	Ctx      context.Context
	Selector interface{}
	Update   interface{}
} {
	var calls []struct {
		Ctx      context.Context
		Selector interface{}
		Update   interface{}
	}
	mock.lockUpdateMany.RLock()
	calls = mock.calls.UpdateMany
	mock.lockUpdateMany.RUnlock()
	return calls
}

// Upsert calls UpsertFunc.
func (mock *MongoCollectionMock) Upsert(ctx context.Context, selector interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
	if mock.UpsertFunc == nil {
		panic("MongoCollectionMock.UpsertFunc: method is nil but MongoCollection.Upsert was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Selector interface{}
		Update   interface{}
	}{
		Ctx:      ctx,
		Selector: selector,
		Update:   update,
	}
	mock.lockUpsert.Lock()
	mock.calls.Upsert = append(mock.calls.Upsert, callInfo)
	mock.lockUpsert.Unlock()
	return mock.UpsertFunc(ctx, selector, update)
}

// UpsertCalls gets all the calls that were made to Upsert.
// Check the length with:
//     len(mockedMongoCollection.UpsertCalls())
func (mock *MongoCollectionMock) UpsertCalls() []struct {
	Ctx      context.Context
	Selector interface{}
	Update   interface{}
} {
	var calls []struct {
		Ctx      context.Context
		Selector interface{}
		Update   interface{}
	}
	mock.lockUpsert.RLock()
	calls = mock.calls.Upsert
	mock.lockUpsert.RUnlock()
	return calls
}

// UpsertById calls UpsertByIdFunc.
func (mock *MongoCollectionMock) UpsertById(ctx context.Context, id interface{}, update interface{}) (*mongodriver.CollectionUpdateResult, error) {
	if mock.UpsertByIdFunc == nil {
		panic("MongoCollectionMock.UpsertByIdFunc: method is nil but MongoCollection.UpsertById was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		ID     interface{}
		Update interface{}
	}{
		Ctx:    ctx,
		ID:     id,
		Update: update,
	}
	mock.lockUpsertById.Lock()
	mock.calls.UpsertById = append(mock.calls.UpsertById, callInfo)
	mock.lockUpsertById.Unlock()
	return mock.UpsertByIdFunc(ctx, id, update)
}

// UpsertByIdCalls gets all the calls that were made to UpsertById.
// Check the length with:
//     len(mockedMongoCollection.UpsertByIdCalls())
func (mock *MongoCollectionMock) UpsertByIdCalls() []struct {
	Ctx    context.Context
	ID     interface{}
	Update interface{}
} {
	var calls []struct {
		Ctx    context.Context
		ID     interface{}
		Update interface{}
	}
	mock.lockUpsertById.RLock()
	calls = mock.calls.UpsertById
	mock.lockUpsertById.RUnlock()
	return calls
}
