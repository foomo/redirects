package redirectstore_test

import (
	"testing"

	storex "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectRequest_GenericTransform(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		in       string
		want     string
		redirect bool
	}{
		{name: "lowercase", in: "/Foo", want: "/foo", redirect: true},
		{name: "trailing slash", in: "/foo/", want: "/foo", redirect: true},
		{name: "keeps query", in: "/Foo/?a=1", want: "/foo?a=1", redirect: true},
		{name: "already normalized", in: "/foo", want: "/foo", redirect: false},
		// open redirect attempts: must never produce a redirect
		{name: "protocol relative", in: "//evil.com/Foo/", want: "//evil.com/Foo/", redirect: false},
		{name: "encoded slash", in: "/%2Fevil.com/Foo", want: "/%2Fevil.com/Foo", redirect: false},
		{name: "backslash", in: "/\\evil.com/Foo", want: "/\\evil.com/Foo", redirect: false},
		{name: "absolute url", in: "https://evil.com/Foo", want: "https://evil.com/Foo", redirect: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, redirect, err := storex.RedirectRequest(tt.in).GenericTransform()
			require.NoError(t, err)
			assert.Equal(t, tt.redirect, redirect)
			assert.Equal(t, tt.want, got.String())
		})
	}
}
