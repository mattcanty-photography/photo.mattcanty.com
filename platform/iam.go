package main

import "encoding/json"

type assumeRolePolicyDocument struct {
	Version   string
	Statement []assumeRolePolicyStatmentEntry
}

type assumeRolePolicyStatmentEntry struct {
	Sid       string
	Effect    string
	Principal assumeRolePolicyStatmentEntryPrincipal
	Action    string
}

type assumeRolePolicyStatmentEntryPrincipal struct {
	Service string
}

type policyDocument struct {
	Version   string
	Statement []policyStatementEntry
}

type policyStatementEntry struct {
	Effect       string
	Action       []string
	Resource     []string
	resourceArgs []interface{}
}

func newPolicyDocumentString(statementEntries ...policyStatementEntry) (string, []interface{}, error) {
	var args []interface{}
	for _, statement := range statementEntries {
		args = append(args, statement.resourceArgs...)
	}
	var doc policyDocument
	doc.Version = "2012-10-17"
	doc.Statement = statementEntries

	byteSlice, err := json.Marshal(&doc)

	return string(byteSlice), args, err
}
