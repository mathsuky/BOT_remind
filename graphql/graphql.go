package graphql

import (
    "context"

    "github.com/hasura/go-graphql-client"
)

type UpdateProjectV2ItemFieldValueInput struct {
    ItemID    graphql.ID `json:"itemId"`
    ProjectID graphql.ID `json:"projectId"`
    FieldID   graphql.ID `json:"fieldId"`
    Value     struct {
        Date graphql.String `json:"date"`
    } `json:"value"`
}

type Mutation struct {
    UpdateProjectV2ItemFieldValue struct {
        ProjectV2Item struct {
            ID graphql.String `graphql:"id"`
        } `graphql:"projectV2Item"`
    } `graphql:"updateProjectV2ItemFieldValue(input: $input)"`
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
                        Id   string
                        Name string
                    } `graphql:"... on ProjectV2Field"`
                }
            } `graphql:"fields(first: 100)"`
        } `graphql:"projectV2(number: $projectNumber)"`
    } `graphql:"user(login: $user)"`
}

func MakeCache(client *graphql.Client) (string, map[int]string, map[string]graphql.ID, error) {
    var info GetProjectBaseInfoQuery
    // キャッシュがない場合はクエリを実行してキャッシュを保存
    err := client.Query(context.Background(), &info, map[string]interface{}{
        "projectNumber": graphql.Int(3),
        "user":          graphql.String("mathsuky"),
    })
    if err != nil {
        return "", nil, nil, err
    }

    projectId := info.User.ProjectV2.Id
    issuesDict := make(map[int]string)
    for _, item := range info.User.ProjectV2.Items.Nodes {
        issuesDict[item.Content.Issue.Number] = item.Id
    }
    fieldsDict := make(map[string]graphql.ID)
    for _, field := range info.User.ProjectV2.Fields.Nodes {
        fieldsDict[field.ProjectV2Field.Name] = graphql.ID(field.ProjectV2Field.Id)
    }

    return projectId, issuesDict, fieldsDict, nil
}