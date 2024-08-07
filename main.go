package main

import (
	"bytes"
	"errors"
	"image"
	"net"
	"net/http"

	"github.com/KononK/resize"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
	"github.com/mattn/go-sixel"
)

const (
	host = "localhost"
	port = "23234"
)

func SixelEncode(url string, width uint) string {
	// download the image
	resp, err := http.Get(url)
	if err != nil {
		log.Error("erroring getting image", "err", err)
		return ""
	}
	defer resp.Body.Close()

	// decode the image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Error("erroring decoding image", "err", err)
		return ""
	}

	// resize image
	m := resize.Resize(width, 0, img, resize.NearestNeighbor)

	// encode the image as sixel and print to stdout
	var buf bytes.Buffer
	sixel.NewEncoder(&buf).Encode(m)
	result := buf.String()

	return result
}

func main() {
	srv, err := wish.NewServer(
		// The address the server will listen to.
		wish.WithAddress(net.JoinHostPort(host, port)),

		// The SSH server need its own keys, this will create a keypair in the
		// given path if it doesn't exist yet.
		// By default, it will create an ED25519 key.
		wish.WithHostKeyPath(".ssh/id_ed25519"),

		// Middlewares do something on a ssh.Session, and then call the next
		// middleware in the stack.
		wish.WithMiddleware(
			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {
					wish.Println(sess, "Hello, world!")
					sixel := SixelEncode("https://emoji.slack-edge.com/T0266FRGM/blob_thumbs_up/1ef9fba2c56e12aa.png", 0)
					wish.Println(sess, sixel)
					next(sess)
				}
			},

			// The last item in the chain is the first to be called.
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	log.Info("Starting SSH server", "host", host, "port", port)
	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		// We ignore ErrServerClosed because it is expected.
		log.Error("Could not start server", "error", err)
	}
}
