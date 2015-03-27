package main // import "github.com/marksteve/xlsx-sms"

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/drone/config"
	"github.com/dustin/randbo"
	"github.com/mozillazg/request"
	"github.com/tealeg/xlsx"
	"gopkg.in/unrolled/render.v1"
)

var (
	addr            = config.String("addr", ":8080")
	chikkaClientId  = config.String("chikka-client-id", "")
	chikkaSecretKey = config.String("chikka-secret-key", "")
	chikkaShortcode = config.String("chikka-shortcode", "")
	r               = render.New()
)

type SMS struct {
	Number  string
	Message string
}

func genId() string {
	p := make([]byte, 4)
	randbo.New().Read(p)
	return fmt.Sprintf("%x", p)
}

func sender(sms <-chan SMS) {
	c := &http.Client{}
	req := request.NewRequest(c)
	for {
		s := <-sms
		req.Data = map[string]string{
			"message_type":  "SEND",
			"mobile_number": s.Number,
			"shortcode":     *chikkaShortcode,
			"message_id":    genId(),
			"message":       s.Message + "\n\n*",
			"client_id":     *chikkaClientId,
			"secret_key":    *chikkaSecretKey,
		}
		log.WithFields(log.Fields{
			"sms": s,
		}).Info("Sending...")
		req.Post("https://post.chikka.com/smsapi/request")
	}
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	r.HTML(w, http.StatusOK, "index", nil)
}

func uploadHandler(sms chan<- SMS) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		f, h, _ := req.FormFile("file")
		defer f.Close()
		tmpFn := filepath.Join(os.TempDir(), h.Filename)
		tmp, _ := os.Create(tmpFn)
		defer tmp.Close()
		io.Copy(tmp, f)
		xl, _ := xlsx.OpenFile(tmpFn)
		numCol := -1
		msgCol := -1
		for _, row := range xl.Sheets[0].Rows {
			if numCol >= 0 && msgCol >= 0 {
				number := row.Cells[numCol].String()
				message := row.Cells[msgCol].String()
				if number != "" && message != "" {
					sms <- SMS{
						Number:  number,
						Message: message,
					}
				}
			} else {
				for x, cell := range row.Cells {
					if strings.ToLower(cell.String()) == "number" {
						numCol = x
					}
					if strings.ToLower(cell.String()) == "message" {
						msgCol = x
					}
				}
			}
		}
		http.Redirect(w, req, "/", 303)
	}
}

func main() {
	config.SetPrefix("XLS_SMS_")
	config.Parse("conf.toml")
	sms := make(chan SMS)
	go sender(sms)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler(sms))
	log.WithFields(log.Fields{
		"addr": *addr,
	}).Info("Starting server...")
	log.Fatal(http.ListenAndServe(*addr, nil))
}
