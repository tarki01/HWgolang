package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
)

type CBRService struct {
	log *logrus.Logger
}

func NewCBRService(log *logrus.Logger) *CBRService {
	return &CBRService{log: log}
}

func (s *CBRService) GetKeyRate() (float64, error) {
	soap := s.buildSOAPRequest()
	body, err := s.sendRequest(soap)
	if err != nil {
		s.log.WithError(err).Error("CBR request failed")
		return 0, err
	}
	rate, err := parseXMLResponse(body)
	if err != nil {
		s.log.WithError(err).Error("CBR response parse failed")
		return 0, err
	}
	rate += 5
	s.log.WithField("rate", rate).Info("CBR key rate fetched")
	return rate, nil
}

func (s *CBRService) buildSOAPRequest() string {
	from := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	to := time.Now().Format("2006-01-02")
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
  <soap12:Body>
    <KeyRate xmlns="http://web.cbr.ru/">
      <fromDate>%s</fromDate>
      <ToDate>%s</ToDate>
    </KeyRate>
  </soap12:Body>
</soap12:Envelope>`, from, to)
}

func (s *CBRService) sendRequest(soap string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST",
		"https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx",
		bytes.NewBufferString(soap))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://web.cbr.ru/KeyRate")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cbr request: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func parseXMLResponse(raw []byte) (float64, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(raw); err != nil {
		return 0, fmt.Errorf("xml parse: %w", err)
	}

	elements := doc.FindElements("//diffgram/KeyRate/KR")
	if len(elements) == 0 {
		return 0, errors.New("rate data not found in response")
	}

	rateEl := elements[0].FindElement("./Rate")
	if rateEl == nil {
		return 0, errors.New("Rate element missing")
	}

	var rate float64
	if _, err := fmt.Sscanf(rateEl.Text(), "%f", &rate); err != nil {
		return 0, fmt.Errorf("rate parse: %w", err)
	}
	return rate, nil
}
