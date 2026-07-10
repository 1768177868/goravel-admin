package middleware

import "github.com/goravel/framework/contracts/http"

type handlerMiddleware struct {
	signature string
	handle    func(http.Context)
}

func newMiddleware(signature string, handle func(http.Context)) http.Middleware {
	return &handlerMiddleware{
		signature: signature,
		handle:    handle,
	}
}

func (m *handlerMiddleware) Handle(ctx http.Context) {
	m.handle(ctx)
}

func (m *handlerMiddleware) Signature() string {
	return m.signature
}
