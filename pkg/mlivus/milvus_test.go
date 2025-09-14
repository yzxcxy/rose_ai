package mlivus

import (
	"context"
	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"testing"
)
import "github.com/milvus-io/milvus/client/v2/milvusclient"

func TestMilvusConnection(t *testing.T) {
	c, err := milvusclient.New(context.Background(), &milvusclient.ClientConfig{
		Address: "localhost:19530",
	})

	if err != nil {
		t.Error(err)
	} else {
		t.Log("Milvus client created successfully:", c)
	}
}

func TestMilvusCreateDatabase(t *testing.T) {
	c, err := milvusclient.New(context.Background(), &milvusclient.ClientConfig{
		Address: "localhost:19530",
	})

	if err != nil {
		t.Error(err)
	} else {
		t.Log("Milvus client created successfully:", c)
	}

	err = c.CreateDatabase(context.Background(), milvusclient.NewCreateDatabaseOption("my_database_1"))
	if err != nil {
		t.Error(err)
	} else {
		t.Log("Milvus database created successfully")
	}
}

func TestMilvusCreateCollection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	milvusAddr := "localhost:19530"
	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: milvusAddr,
		DBName:  "my_database_1",
	})
	if err != nil {
		t.Error(err)
	} else {
		t.Log("Milvus client created successfully:", client)
	}
	defer client.Close(ctx)

	schema := entity.NewSchema().WithDynamicFieldEnabled(true).
		WithField(entity.NewField().WithName("my_id").WithIsAutoID(false).WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true)).
		WithField(entity.NewField().WithName("my_vector").WithDataType(entity.FieldTypeFloatVector).WithDim(5)).
		WithField(entity.NewField().WithName("my_varchar").WithDataType(entity.FieldTypeVarChar).WithMaxLength(512))

	err = client.CreateCollection(ctx, milvusclient.NewCreateCollectionOption("customized_setup_2", schema))
	if err != nil {
		t.Error(err)
	} else {
		t.Log("Milvus collection created successfully")
	}

}

func TestMilvusInsertData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: "localhost:19530",
		DBName:  "my_database_1", // 可选，按需填写
	})
	if err != nil {
		t.Error(err)
	}
	defer client.Close(ctx)

	dynamicColumn := column.NewColumnString("color", []string{
		"pink_8682", "red_7025", "orange_6781", "pink_9298", "red_4794",
		"yellow_4222", "red_9392", "grey_8510", "white_9381", "purple_4976",
	})

	myVarcharColumn := column.NewColumnString("my_varchar", []string{
		"this is a test string 0", "this is a test string 1", "this is a test string 2", "this is a test string 3",
		"this is a test string 4", "this is a test string 5", "this is a test string 6", "this is a test string 7",
		"this is a test string 8", "this is a test string 9",
	})

	_, err = client.Insert(ctx, milvusclient.NewColumnBasedInsertOption("customized_setup_2").
		WithInt64Column("my_id", []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}).
		WithFloatVectorColumn("my_vector", 5, [][]float32{
			{0.35803763, -0.60234957, 0.18414012, -0.26286205, 0.90294384},
			{0.19886812, 0.06023560, 0.69769630, 0.26144745, 0.83872948},
			{0.43742130, -0.55975025, 0.64578876, 0.78940589, 0.20785793},
			{0.31720052, 0.97190447, -0.36981146, -0.48608945, 0.95791889},
			{0.44523495, -0.87570269, 0.82207794, 0.46406290, 0.30337481},
			{0.98582513, -0.81446515, 0.62992670, 0.12069069, -0.14462777},
			{0.83719777, -0.01576436, -0.31062937, -0.56266695, -0.89849476},
			{-0.33445148, -0.25671350, 0.89875397, 0.94029958, 0.53780649},
			{0.39524717, 0.40002572, -0.58905073, -0.86505022, -0.61403607},
			{0.57182804, 0.24070317, -0.37379134, -0.06726932, -0.69805316},
		}).
		WithVarcharColumn("my_varchar", myVarcharColumn.Data()).
		WithColumns(dynamicColumn),
	)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("Milvus data inserted successfully")
	}
}

func TestMilvusQueryData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: "localhost:19530",
		DBName:  "my_database_1", // 可选，按需填写
	})
	if err != nil {
		t.Error(err)
	}
	defer client.Close(ctx)

	// 创建索引
	hnswIndex := index.NewHNSWIndex(entity.COSINE, 16, 200)
	indexTask, err := client.CreateIndex(ctx, milvusclient.NewCreateIndexOption("customized_setup_2", "my_vector", hnswIndex))
	if err != nil {
		t.Error(err)
	} else {
		t.Log("Index creation started:", indexTask)
	}

	// 等待装载
	loadTask, err := client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption("customized_setup_2"))
	if err != nil {
		t.Error(err)
	} else {
		t.Log("Collection loading started:", loadTask)
	}

	// sync wait collection to be loaded
	err = loadTask.Await(ctx)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("Collection loaded successfully")
	}

	state, err := client.GetLoadState(ctx, milvusclient.NewGetLoadStateOption("customized_setup_2"))
	if err != nil {
		t.Error(err)
	} else {
		t.Log("Collection load state:", state)
	}

	// 查询数据
	queryVector := []float32{0.35803763, -0.60234957, 0.18414012, -0.26286205, 0.90294384}

	resultSets, err := client.Search(ctx, milvusclient.NewSearchOption(
		"customized_setup_2", // collectionName
		3,                    // limit
		[]entity.Vector{entity.FloatVector(queryVector)},
	).WithANNSField("my_vector").WithOutputFields([]string{"my_id", "my_varchar"}...).
		WithConsistencyLevel(entity.ClStrong))

	if err != nil {
		t.Fatal("Milvus search failed:", err)
	}

	for _, resultSet := range resultSets {
		t.Log("IDs:", resultSet.IDs.FieldData().GetScalars())
		t.Log("Scores:", resultSet.Scores)
		t.Log("my_varchar:", resultSet.GetColumn("my_varchar").FieldData().GetScalars().String())
	}
}
