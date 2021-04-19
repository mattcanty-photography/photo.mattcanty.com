package helpers

import "encoding/json"

type AssumeRolePolicyDocument struct {
	Version   string
	Statement []AssumeRolePolicyStatmentEntry
}

type AssumeRolePolicyStatmentEntry struct {
	Sid       string
	Effect    string
	Principal AssumeRolePolicyStatmentEntryPrincipal
	Action    string
}

type AssumeRolePolicyStatmentEntryPrincipal struct {
	Service string
}

type PolicyDocument struct {
	Version   string
	Statement []PolicyStatementEntry
}

type PolicyStatementEntry struct {
	Effect       string
	Action       []string
	Resource     []string
	resourceArgs []interface{}
}

func NewPolicyDocumentString(statementEntries ...PolicyStatementEntry) (string, []interface{}, error) {
	var args []interface{}
	for _, statement := range statementEntries {
		args = append(args, statement.resourceArgs...)
	}
	var doc PolicyDocument
	doc.Version = "2012-10-17"
	doc.Statement = statementEntries

	byteSlice, err := json.Marshal(&doc)

	return string(byteSlice), args, err
}
