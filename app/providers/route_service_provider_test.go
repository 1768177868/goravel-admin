package providers

import "testing"

func TestIsBadRequestBodyPanicWithStack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		msg   string
		stack string
		want  bool
	}{
		{
			name: "multipart parse error message",
			msg:  `parse multipart form error: malformed MIME header: missing colon: "------WebKitFormBoundaryx123--"`,
			want: true,
		},
		{
			name: "malformed mime header",
			msg:  "malformed MIME header: missing colon",
			want: true,
		},
		{
			name:  "nil pointer in getHttpBody",
			msg:   "runtime error: invalid memory address or nil pointer dereference",
			stack: "goravel/gin@v1.18.0/context_request.go:660\ngin.getHttpBody\ngin.NewContextRequest",
			want:  true,
		},
		{
			name:  "nil pointer in app code is not client error",
			msg:   "runtime error: invalid memory address or nil pointer dereference",
			stack: "goravel/app/services/foo.go:12\nservices.(*Foo).Bar",
			want:  false,
		},
		{
			name: "unrelated panic message",
			msg:  "something went wrong",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isBadRequestBodyPanicWithStack(tt.msg, tt.stack)
			if got != tt.want {
				t.Fatalf("isBadRequestBodyPanicWithStack(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}
