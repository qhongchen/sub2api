package repository

import (
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestBuildOpsErrorLogsWhere_QueryUsesQualifiedColumns(t *testing.T) {
	filter := &service.OpsErrorLogFilter{
		Query: "ACCESS_DENIED",
	}

	where, args := buildOpsErrorLogsWhere(filter)
	if where == "" {
		t.Fatalf("where should not be empty")
	}
	if len(args) != 1 {
		t.Fatalf("args len = %d, want 1", len(args))
	}
	if !strings.Contains(where, "e.request_id ILIKE $") {
		t.Fatalf("where should include qualified request_id condition: %s", where)
	}
	if !strings.Contains(where, "e.client_request_id ILIKE $") {
		t.Fatalf("where should include qualified client_request_id condition: %s", where)
	}
	if !strings.Contains(where, "e.error_message ILIKE $") {
		t.Fatalf("where should include qualified error_message condition: %s", where)
	}
}

func TestBuildOpsErrorLogsWhere_UserQueryUsesExistsSubquery(t *testing.T) {
	filter := &service.OpsErrorLogFilter{
		UserQuery: "admin@",
	}

	where, args := buildOpsErrorLogsWhere(filter)
	if where == "" {
		t.Fatalf("where should not be empty")
	}
	if len(args) != 1 {
		t.Fatalf("args len = %d, want 1", len(args))
	}
	if !strings.Contains(where, "EXISTS (SELECT 1 FROM users u WHERE u.id = e.user_id AND u.email ILIKE $") {
		t.Fatalf("where should include EXISTS user email condition: %s", where)
	}
}

func TestBuildOpsErrorLogsWhere_StatusCodeDefaultsToClientScope(t *testing.T) {
	filter := &service.OpsErrorLogFilter{
		StatusCodes: []int{502},
	}

	where, args := buildOpsErrorLogsWhere(filter)

	if len(args) != 1 {
		t.Fatalf("args len = %d, want 1", len(args))
	}
	if !strings.Contains(where, "COALESCE(e.status_code, 0) = ANY($") {
		t.Fatalf("where should filter by client status_code by default: %s", where)
	}
	if strings.Contains(where, "COALESCE(e.upstream_status_code, e.status_code, 0) = ANY($") {
		t.Fatalf("where should not use upstream status_code in client scope: %s", where)
	}
}

func TestBuildOpsErrorLogsWhere_StatusCodeCanUseUpstreamScope(t *testing.T) {
	filter := &service.OpsErrorLogFilter{
		StatusCodes:     []int{401},
		StatusCodeScope: "upstream",
		Phase:           "upstream",
	}

	where, args := buildOpsErrorLogsWhere(filter)

	if len(args) != 2 {
		t.Fatalf("args len = %d, want 2", len(args))
	}
	if !strings.Contains(where, "COALESCE(e.upstream_status_code, e.status_code, 0) = ANY($") {
		t.Fatalf("where should filter by upstream status_code in upstream scope: %s", where)
	}
}

func TestBuildOpsErrorLogsWhere_StatusCodesOtherUsesSelectedStatusScope(t *testing.T) {
	t.Run("client scope", func(t *testing.T) {
		filter := &service.OpsErrorLogFilter{StatusCodesOther: true}

		where, args := buildOpsErrorLogsWhere(filter)

		if len(args) != 1 {
			t.Fatalf("args len = %d, want 1", len(args))
		}
		if !strings.Contains(where, "NOT (COALESCE(e.status_code, 0) = ANY($") {
			t.Fatalf("where should exclude common client status codes: %s", where)
		}
	})

	t.Run("upstream scope", func(t *testing.T) {
		filter := &service.OpsErrorLogFilter{
			StatusCodesOther: true,
			StatusCodeScope:  "upstream",
			Phase:            "upstream",
		}

		where, args := buildOpsErrorLogsWhere(filter)

		if len(args) != 2 {
			t.Fatalf("args len = %d, want 2", len(args))
		}
		if !strings.Contains(where, "NOT (COALESCE(e.upstream_status_code, e.status_code, 0) = ANY($") {
			t.Fatalf("where should exclude common upstream status codes: %s", where)
		}
	})
}
