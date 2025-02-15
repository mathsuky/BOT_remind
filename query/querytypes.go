package query

import (
	"time"

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

type AddAssigneesToAssignableInput struct {
	AssignableId graphql.ID   `json:"assignableId"`
	AssigneeIds  []graphql.ID `json:"assigneeIds"`
}

type AddAssigneeToAssignableMutation struct {
	AddAssigneesToAssignable struct {
		Assignable struct {
			Issue struct {
				Title string
			} `graphql:"... on Issue"`
		} `graphql:"assignable"`
	} `graphql:"addAssigneesToAssignable(input: $input)"`
}

type GetIssueIdFromRepositoryQuery struct {
	Repository struct {
		Issue struct {
			Id string
		} `graphql:"issue(number: $issueNumber)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

type GetUserIdQuery struct {
	User struct {
		Id string
	} `graphql:"user(login: $login)"`
}

type GetOrganizationProjectBaseInfoQuery struct {
	Organization struct {
		ProjectV2 struct {
			Id     string
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
					} `graphql:"... on ProjectV2SingleSelectField"`
				}
			} `graphql:"fields(first: 100)"`
		} `graphql:"projectV2(number: $projectNumber)"`
	} `graphql:"organization(login: $organization)"`
}

type GetOrganizationProjectItemsQuery struct {
	Organization struct {
		ProjectV2 struct {
			Items struct {
				Nodes []struct {
					Id      string
					Content struct {
						Issue struct {
							Number int
						} `graphql:"... on Issue"`
					}
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
			} `graphql:"items(first: $first, after: $after)"`
		} `graphql:"projectV2(number: $projectNumber)"`
	} `graphql:"organization(login: $organization)"`
}

type GetIssueFieldsQuery struct {
	Organization struct {
		ProjectV2 struct {
			Items struct {
				Nodes []struct {
					Content struct {
						Issue struct {
							Number int
						} `graphql:"... on Issue"`
					}
					Deadline struct {
						ProjectV2ItemFieldDateValue struct {
							Date string
						} `graphql:"... on ProjectV2ItemFieldDateValue"`
					} `graphql:"Deadline: fieldValueByName(name: \"目標\")"`
					Status struct {
						ProjectV2ItemFieldSingleSelectValue struct {
							Name string
						} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
					} `graphql:"status: fieldValueByName(name: \"Status\")"`
					TraqID struct {
						ProjectV2ItemFieldTextValue struct {
							Text string
						} `graphql:"... on ProjectV2ItemFieldTextValue"`
					} `graphql:"traQID: fieldValueByName(name: \"traQID\")"`
				}
			} `graphql:"items(first: 100)"`
		} `graphql:"projectV2(number: $projectNumber)"`
	} `graphql:"organization(login: $organization)"`
}

type ProjectV2SingleSelectFieldOption struct {
	Id   string `graphql:"id"`
	Name string `graphql:"name"`
}

type IssueDetail struct {
	IssueNum int
	Assignee string
	Deadline time.Time
	Status   string
}
