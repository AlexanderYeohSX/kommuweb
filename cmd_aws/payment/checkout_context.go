package main

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// CheckoutContextPayload is the full fulfillment payload stored outside Razorpay notes.
type CheckoutContextPayload struct {
	Kind      string `json:"kind"` // order | subscription
	CartRef   string `json:"cart_ref"`
	DevTest   bool   `json:"dev_test"`
	Notes     notes  `json:"notes"`
	RawItems  string `json:"raw_items,omitempty"`
	CreatedAt string `json:"created_at"`
}

type checkoutContextRecord struct {
	EntityID  string `dynamodbav:"entity_id"`
	CartRef   string `dynamodbav:"cart_ref"`
	Kind      string `dynamodbav:"kind"`
	Payload   string `dynamodbav:"payload"`
	CreatedAt string `dynamodbav:"created_at"`
	TTL       int64  `dynamodbav:"ttl"`
}

var (
	ddbClient     *dynamodb.Client
	ddbClientOnce sync.Once
	ddbClientErr  error
)

func getDynamoClient() (*dynamodb.Client, error) {
	ddbClientOnce.Do(func() {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			ddbClientErr = err
			return
		}
		ddbClient = dynamodb.NewFromConfig(cfg)
	})
	return ddbClient, ddbClientErr
}

func putCheckoutContext(entityID string, payload CheckoutContextPayload) error {
	table := checkoutContextTable()
	if table == "" || entityID == "" {
		return nil
	}
	client, err := getDynamoClient()
	if err != nil || client == nil {
		curlecLog("checkout_context.put_client_error", map[string]string{"error": errString(err)})
		return err
	}
	if payload.CreatedAt == "" {
		payload.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	rec := checkoutContextRecord{
		EntityID:  entityID,
		CartRef:   payload.CartRef,
		Kind:      payload.Kind,
		Payload:   string(raw),
		CreatedAt: payload.CreatedAt,
		TTL:       time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	item, err := attributevalue.MarshalMap(rec)
	if err != nil {
		return err
	}
	_, err = client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item,
	})
	if err != nil {
		curlecLog("checkout_context.put_error", map[string]string{"entity_id": entityID, "error": err.Error()})
		return err
	}
	curlecLog("checkout_context.put_ok", map[string]string{"entity_id": entityID, "cart_ref": payload.CartRef})
	return nil
}

func getCheckoutContext(entityID string) (*CheckoutContextPayload, error) {
	table := checkoutContextTable()
	if table == "" || entityID == "" {
		return nil, nil
	}
	client, err := getDynamoClient()
	if err != nil || client == nil {
		return nil, err
	}
	out, err := client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"entity_id": &types.AttributeValueMemberS{Value: entityID},
		},
	})
	if err != nil {
		curlecLog("checkout_context.get_error", map[string]string{"entity_id": entityID, "error": err.Error()})
		return nil, err
	}
	if out.Item == nil {
		return nil, nil
	}
	var rec checkoutContextRecord
	if err := attributevalue.UnmarshalMap(out.Item, &rec); err != nil {
		return nil, err
	}
	var payload CheckoutContextPayload
	if err := json.Unmarshal([]byte(rec.Payload), &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// hydrateNotesFromContext merges DynamoDB checkout context over Razorpay notes.
func hydrateNotesFromContext(entityID string, fromRazorpay notes) notes {
	ctx, err := getCheckoutContext(entityID)
	if err != nil || ctx == nil {
		return fromRazorpay
	}
	merged := mergeNotes(fromRazorpay, ctx.Notes)
	if ctx.RawItems != "" && merged.Items == "" && merged.Items1 == "" {
		merged.Items = ctx.RawItems
	}
	if ctx.DevTest {
		merged.DevTest = "1"
	}
	curlecLog("checkout_context.hydrate_ok", map[string]string{"entity_id": entityID, "cart_ref": merged.CartRef})
	return merged
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func dynamoDisabledForTests() bool {
	return os.Getenv("CHECKOUT_CONTEXT_DISABLE") == "1"
}
