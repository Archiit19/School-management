package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Archiit19/School-management/user-service/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// UserProfileDoc is stored in DynamoDB keyed by user_id.
type UserProfileDoc struct {
	UserID    string                 `dynamodbav:"user_id"`
	RoleID    string                 `dynamodbav:"role_id"`
	SchoolID  string                 `dynamodbav:"school_id"`
	Data      map[string]interface{} `dynamodbav:"data"`
	UpdatedAt string                 `dynamodbav:"updated_at"`
}

type ProfileRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewProfileRepository(cfg *config.Config) (*ProfileRepository, error) {
	ctx := context.Background()
	var opts []func(*awsconfig.LoadOptions) error

	if cfg.DynamoEndpoint != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("local", "local", ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
		if cfg.DynamoEndpoint != "" {
			o.BaseEndpoint = aws.String(cfg.DynamoEndpoint)
		}
	})

	repo := &ProfileRepository{client: client, tableName: cfg.DynamoTable}
	if err := repo.ensureTable(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *ProfileRepository) ensureTable(ctx context.Context) error {
	_, err := r.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(r.tableName),
	})
	if err == nil {
		return nil
	}

	_, err = r.client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(r.tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("user_id"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("user_id"), KeyType: types.KeyTypeHash},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		return fmt.Errorf("create dynamodb table %s: %w", r.tableName, err)
	}

	waiter := dynamodb.NewTableExistsWaiter(r.client)
	return waiter.Wait(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(r.tableName)}, 30*time.Second)
}

func (r *ProfileRepository) Save(userID uuid.UUID, roleID, schoolID uuid.UUID, data map[string]interface{}) error {
	doc := UserProfileDoc{
		UserID:    userID.String(),
		RoleID:    roleID.String(),
		SchoolID:  schoolID.String(),
		Data:      data,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	item, err := attributevalue.MarshalMap(doc)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}
	_, err = r.client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	return err
}

func (r *ProfileRepository) Get(userID uuid.UUID) (*UserProfileDoc, error) {
	out, err := r.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID.String()},
		},
	})
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, errors.New("profile not found")
	}
	var doc UserProfileDoc
	if err := attributevalue.UnmarshalMap(out.Item, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *ProfileRepository) Delete(userID uuid.UUID) error {
	_, err := r.client.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID.String()},
		},
	})
	return err
}

func (r *ProfileRepository) BatchGet(userIDs []uuid.UUID) (map[uuid.UUID]*UserProfileDoc, error) {
	result := make(map[uuid.UUID]*UserProfileDoc)
	if len(userIDs) == 0 {
		return result, nil
	}

	const batchSize = 100
	for i := 0; i < len(userIDs); i += batchSize {
		end := i + batchSize
		if end > len(userIDs) {
			end = len(userIDs)
		}
		keys := make([]map[string]types.AttributeValue, 0, end-i)
		for _, id := range userIDs[i:end] {
			keys = append(keys, map[string]types.AttributeValue{
				"user_id": &types.AttributeValueMemberS{Value: id.String()},
			})
		}
		out, err := r.client.BatchGetItem(context.Background(), &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				r.tableName: {Keys: keys},
			},
		})
		if err != nil {
			return nil, err
		}
		for _, item := range out.Responses[r.tableName] {
			var doc UserProfileDoc
			if err := attributevalue.UnmarshalMap(item, &doc); err != nil {
				continue
			}
			id, _ := uuid.Parse(doc.UserID)
			result[id] = &doc
		}
	}
	return result, nil
}

// ParseChildrenIDs normalizes the children list stored in role profile data.
func ParseChildrenIDs(raw interface{}) []string {
	if raw == nil {
		return []string{}
	}
	switch v := raw.(type) {
	case []string:
		return append([]string{}, v...)
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			s := strings.TrimSpace(fmt.Sprint(item))
			if s != "" && s != "<nil>" {
				out = append(out, s)
			}
		}
		return out
	default:
		s := strings.TrimSpace(fmt.Sprint(v))
		if s == "" || s == "<nil>" {
			return []string{}
		}
		return []string{s}
	}
}

func (r *ProfileRepository) HasChild(parentID, childID uuid.UUID) (bool, error) {
	doc, err := r.Get(parentID)
	if err != nil {
		return false, err
	}
	childStr := childID.String()
	for _, id := range ParseChildrenIDs(doc.Data["children"]) {
		if id == childStr {
			return true, nil
		}
	}
	return false, nil
}

func (r *ProfileRepository) AppendChild(parentID, childID uuid.UUID) error {
	doc, err := r.Get(parentID)
	if err != nil {
		return err
	}
	if doc.Data == nil {
		doc.Data = map[string]interface{}{}
	}
	children := ParseChildrenIDs(doc.Data["children"])
	childStr := childID.String()
	for _, id := range children {
		if id == childStr {
			return nil
		}
	}
	children = append(children, childStr)
	list := make([]interface{}, len(children))
	for i, id := range children {
		list[i] = id
	}
	doc.Data["children"] = list
	schoolID, _ := uuid.Parse(doc.SchoolID)
	roleID, _ := uuid.Parse(doc.RoleID)
	return r.Save(parentID, roleID, schoolID, doc.Data)
}

func (r *ProfileRepository) RemoveChild(parentID, childID uuid.UUID) error {
	doc, err := r.Get(parentID)
	if err != nil {
		return err
	}
	if doc.Data == nil {
		return nil
	}
	childStr := childID.String()
	children := ParseChildrenIDs(doc.Data["children"])
	next := make([]string, 0, len(children))
	for _, id := range children {
		if id != childStr {
			next = append(next, id)
		}
	}
	if len(next) == len(children) {
		return nil
	}
	list := make([]interface{}, len(next))
	for i, id := range next {
		list[i] = id
	}
	doc.Data["children"] = list
	schoolID, _ := uuid.Parse(doc.SchoolID)
	roleID, _ := uuid.Parse(doc.RoleID)
	return r.Save(parentID, roleID, schoolID, doc.Data)
}
