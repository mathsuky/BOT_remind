package query

import (
	"github.com/hasura/go-graphql-client"
)

type FieldValue map[string]interface{}

type UpdateProjectV2ItemFieldValueInput struct {
	ItemID    graphql.ID `json:"itemId"`
	ProjectID graphql.ID `json:"projectId"`
	FieldID   graphql.ID `json:"fieldId"`
	Value     FieldValue `json:"value"`
}

type UpdateProjectV2ItemFieldValue struct {
	UpdateProjectV2ItemFieldValue struct {
		ProjectV2Item struct {
			ID graphql.String `graphql:"id"`
		} `graphql:"projectV2Item"`
	} `graphql:"updateProjectV2ItemFieldValue(input: $input)"`
}

type UpdateRelatedIssueDeadlineMutation struct {
	UpdateGoal struct {
		ProjectV2Item struct {
			ID graphql.String `graphql:"id"`
		} `graphql:"projectV2Item"`
	} `graphql:"updateGoal:updateProjectV2ItemFieldValue(input: $input1)"`
	UpdateStart struct {
		ProjectV2Item struct {
			ID graphql.String `graphql:"id"`
		} `graphql:"projectV2Item"`
	} `graphql:"updateStart:updateProjectV2ItemFieldValue(input: $input2)"`
}

type GetProjectBaseInfoQuery struct {
	User struct {
		ProjectV2 struct {
			Id    string
			Items struct {
				Nodes []struct {
					Id      string
					Content struct {
						Issue struct {
							Number int
						} `graphql:"... on Issue"`
					}
				}
			} `graphql:"items(first: 100)"`
			Fields struct {
				Nodes []struct {
					ProjectV2Field struct {
						Id       string
						Name     string
						DataType string `graphql:"dataType"`
					} `graphql:"... on ProjectV2Field"`
					ProjectV2SingleSelectField struct {
						Id      string
						Name    string
						Options []struct {
							Id   string
							Name string
						} `graphql:"ProjectV2SingleSelectFieldOption`
					} `graphql:"... on ProjectV2SingleSelectField"` // これがいるかわからない
				}
			} `graphql:"fields(first: 100)"`
		} `graphql:"projectV2(number: $projectNumber)"`
	} `graphql:"user(login: $user)"`
}

type ProjectV2SingleSelectFieldOption struct {
	Id   string `graphql:"id"`
	Name string `graphql:"name"`
}
