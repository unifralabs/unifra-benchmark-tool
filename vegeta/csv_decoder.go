package vegeta

import (
	"bufio"
	"encoding/base64"
	"encoding/csv"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	Attack    string        `json:"attack"`
	Seq       uint64        `json:"seq"`
	Code      uint16        `json:"code"`
	Timestamp time.Time     `json:"timestamp"`
	Latency   time.Duration `json:"latency"`
	BytesOut  uint64        `json:"bytes_out"`
	BytesIn   uint64        `json:"bytes_in"`
	Error     string        `json:"error"`
	Body      []byte        `json:"body"`
	Method    string        `json:"method"`
	URL       string        `json:"url"`
	Headers   http.Header   `json:"headers"`
}

// A Decoder decodes a Result and returns an error in case of failure.
type Decoder func(*Result) error

// NewCSVDecoder returns a Decoder that decodes CSV encoded Results.
func NewCSVDecoder(r io.Reader) Decoder {
	dec := csv.NewReader(r)
	dec.FieldsPerRecord = 12
	dec.TrimLeadingSpace = true

	return func(r *Result) error {
		rec, err := dec.Read()
		if err != nil {
			return err
		}

		ts, err := strconv.ParseInt(rec[0], 10, 64)
		if err != nil {
			return err
		}
		r.Timestamp = time.Unix(0, ts)

		code, err := strconv.ParseUint(rec[1], 10, 16)
		if err != nil {
			return err
		}
		r.Code = uint16(code)

		latency, err := strconv.ParseInt(rec[2], 10, 64)
		if err != nil {
			return err
		}
		r.Latency = time.Duration(latency)

		if r.BytesOut, err = strconv.ParseUint(rec[3], 10, 64); err != nil {
			return err
		}

		if r.BytesIn, err = strconv.ParseUint(rec[4], 10, 64); err != nil {
			return err
		}

		r.Error = rec[5]
		if r.Body, err = base64.StdEncoding.DecodeString(rec[6]); err != nil {
			return err
		}

		r.Attack = rec[7]
		if r.Seq, err = strconv.ParseUint(rec[8], 10, 64); err != nil {
			return err
		}

		r.Method = rec[9]
		r.URL = rec[10]

		if rec[11] != "" {
			pr := textproto.NewReader(bufio.NewReader(
				base64.NewDecoder(base64.StdEncoding, strings.NewReader(rec[11]))))
			hdr, err := pr.ReadMIMEHeader()
			if err != nil {
				return err
			}
			r.Headers = http.Header(hdr)
		}

		return err
	}
}
