package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"

	"github.com/yankeguo/rg"
)

type RegexpList []*regexp.Regexp

func NewRegexpList(items []string) (rl RegexpList, err error) {
	defer rg.Guard(&err)

	rl = make(RegexpList, 0, len(items))

	for _, item := range items {
		rl = append(rl, rg.Must(regexp.Compile(item)))
	}
	return
}

func (rl RegexpList) Match(s string) bool {
	for _, re := range rl {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

type WebhookCall struct {
	Source string
	Target string
}

func (c WebhookCall) Inject(m map[string]string) map[string]string {
	return expandMap(m, map[string]string{
		"SOURCE_IMAGE": c.Source,
		"TARGET_IMAGE": c.Target,
	})
}

type Replacer struct {
	opts Options

	matchNamespace RegexpList
	matchImage     RegexpList
	webhookCalls   map[WebhookCall]struct{}
}

func NewReplacer(opts Options) (m *Replacer, err error) {
	defer rg.Guard(&err)

	m = &Replacer{
		opts:           opts,
		matchNamespace: rg.Must(NewRegexpList(opts.ImageAutoMapping.Match.Namespaces)),
		matchImage:     rg.Must(NewRegexpList(opts.ImageAutoMapping.Match.Images)),
		webhookCalls:   make(map[WebhookCall]struct{}),
	}
	return
}

func (m *Replacer) Lookup(namespace string, image string) (newImage string, ok bool) {
	image = standardizeImage(image)

	// static mappings
	if m.opts.ImageMappings != nil {
		if newImage, ok = m.opts.ImageMappings[image]; ok {
			return
		}
	}

	// dynamic mappings
	if m.matchNamespace.Match(namespace) && m.matchImage.Match(image) {
		newImageBase := flattenImage(image)

		newImage = path.Join(m.opts.ImageAutoMapping.Registry, newImageBase)
		ok = true

		wi := WebhookCall{
			Source: image,
		}

		if m.opts.ImageAutoMapping.Webhook.Override.Registry == "" {
			wi.Target = newImage
		} else {
			wi.Target = path.Join(m.opts.ImageAutoMapping.Webhook.Override.Registry, newImageBase)
		}

		m.webhookCalls[wi] = struct{}{}
	}
	return
}

func (m *Replacer) invokeWebhook(call WebhookCall) {
	var err error
	defer func() {
		if err == nil {
			return
		}
		log.Println("Error invoking webhook:", err.Error(), "for", call.Source, "->", call.Target)
	}()
	defer rg.Guard(&err)

	wc := m.opts.ImageAutoMapping.Webhook

	if wc.URL == "" {
		return
	}

	webhookURL := rg.Must(url.Parse(wc.URL))
	q := webhookURL.Query()
	for k, v := range call.Inject(wc.Query) {
		q.Set(k, v)
	}
	webhookURL.RawQuery = q.Encode()

	headers := call.Inject(wc.Headers)

	if wc.JSON != nil {
		headers["Content-Type"] = "application/json"
	} else if wc.Form != nil {
		headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	method := http.MethodGet
	if wc.Form != nil || wc.JSON != nil {
		method = http.MethodPost
	}

	var body io.Reader

	if wc.JSON != nil {
		body = bytes.NewReader(rg.Must(json.Marshal(call.Inject(wc.JSON))))
	} else if wc.Form != nil {
		q := url.Values{}
		for k, v := range call.Inject(wc.Form) {
			q.Set(k, v)
		}
		body = bytes.NewReader([]byte(q.Encode()))
	}

	req := rg.Must(http.NewRequest(method, webhookURL.String(), body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res := rg.Must(http.DefaultClient.Do(req))
	defer res.Body.Close()

	buf := rg.Must(io.ReadAll(req.Body))

	log.Println("Webhook response:", res.StatusCode, string(buf))
}

func (m *Replacer) InvokeWebhook() {
	for call := range m.webhookCalls {
		m.invokeWebhook(call)
	}
	m.webhookCalls = make(map[WebhookCall]struct{})
}
