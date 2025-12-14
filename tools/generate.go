// Package tools contains go:generate directives for code generation.
//
// This file is used to trigger code generation from OpenAPI specifications.
// Run `go generate ./tools` to regenerate API client code.
//
// Note: Some operations are excluded due to oapi-codegen limitations with deep schema references.
// Excluded operations: upsert_by_code endpoints (account_items, items, sections, partners, segment_tags)
// See internal/gen/README.md for details on the reference depth issue and workarounds.
package tools

//go:generate oapi-codegen -package gen -generate types,client -exclude-operation-ids api/v1/account_items#upsert_by_code,api/v1/items#upsert_by_code,api/v1/sections#upsert_by_code,upsert_segment_tag,api/v1/partners#upsert_by_code -o ../internal/gen/client.gen.go ../api/openapi.json
