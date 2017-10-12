package mocks

type OnlyWriter struct{}

func (o *OnlyWriter) Write(p []byte) (n int, err error) {
	return 0, nil
}

type OnlyReader struct{}

func (o *OnlyReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

type WriteCloser struct {
	IsClosed bool
}

func (o *WriteCloser) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (o *WriteCloser) Close() error {
	o.IsClosed = true
	return nil
}

type ReadCloser struct {
	IsClosed bool
}

func (o *ReadCloser) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (o *ReadCloser) Close() error {
	o.IsClosed = true
	return nil
}
