// Code generated by ffjson <https://github.com/pquerna/ffjson>. DO NOT EDIT.
// source: geoip.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	fflib "github.com/pquerna/ffjson/fflib/v1"
)

// MarshalJSON marshal bytes to json - template
func (j *IfconfigClient) MarshalJSON() ([]byte, error) {
	var buf fflib.Buffer
	if j == nil {
		buf.WriteString("null")
		return buf.Bytes(), nil
	}
	err := j.MarshalJSONBuf(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalJSONBuf marshal buff to json - template
func (j *IfconfigClient) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if j == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	if j.URL != nil {
		/* Struct fall back. type=url.URL kind=struct */
		buf.WriteString(`{"URL":`)
		err = buf.Encode(j.URL)
		if err != nil {
			return err
		}
	} else {
		buf.WriteString(`{"URL":null`)
	}
	buf.WriteByte('}')
	return nil
}

const (
	ffjtIfconfigClientbase = iota
	ffjtIfconfigClientnosuchkey

	ffjtIfconfigClientURL
)

var ffjKeyIfconfigClientURL = []byte("URL")

// UnmarshalJSON umarshall json - template of ffjson
func (j *IfconfigClient) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return j.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

// UnmarshalJSONFFLexer fast json unmarshall - template ffjson
func (j *IfconfigClient) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error
	currentKey := ffjtIfconfigClientbase
	_ = currentKey
	tok := fflib.FFTok_init
	wantedTok := fflib.FFTok_init

mainparse:
	for {
		tok = fs.Scan()
		//	println(fmt.Sprintf("debug: tok: %v  state: %v", tok, state))
		if tok == fflib.FFTok_error {
			goto tokerror
		}

		switch state {

		case fflib.FFParse_map_start:
			if tok != fflib.FFTok_left_bracket {
				wantedTok = fflib.FFTok_left_bracket
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_key
			continue

		case fflib.FFParse_after_value:
			if tok == fflib.FFTok_comma {
				state = fflib.FFParse_want_key
			} else if tok == fflib.FFTok_right_bracket {
				goto done
			} else {
				wantedTok = fflib.FFTok_comma
				goto wrongtokenerror
			}

		case fflib.FFParse_want_key:
			// json {} ended. goto exit. woo.
			if tok == fflib.FFTok_right_bracket {
				goto done
			}
			if tok != fflib.FFTok_string {
				wantedTok = fflib.FFTok_string
				goto wrongtokenerror
			}

			kn := fs.Output.Bytes()
			if len(kn) <= 0 {
				// "" case. hrm.
				currentKey = ffjtIfconfigClientnosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'U':

					if bytes.Equal(ffjKeyIfconfigClientURL, kn) {
						currentKey = ffjtIfconfigClientURL
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.SimpleLetterEqualFold(ffjKeyIfconfigClientURL, kn) {
					currentKey = ffjtIfconfigClientURL
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffjtIfconfigClientnosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			}

		case fflib.FFParse_want_colon:
			if tok != fflib.FFTok_colon {
				wantedTok = fflib.FFTok_colon
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_value
			continue
		case fflib.FFParse_want_value:

			if tok == fflib.FFTok_left_brace || tok == fflib.FFTok_left_bracket || tok == fflib.FFTok_integer || tok == fflib.FFTok_double || tok == fflib.FFTok_string || tok == fflib.FFTok_bool || tok == fflib.FFTok_null {
				switch currentKey {

				case ffjtIfconfigClientURL:
					goto handle_URL

				case ffjtIfconfigClientnosuchkey:
					err = fs.SkipField(tok)
					if err != nil {
						return fs.WrapErr(err)
					}
					state = fflib.FFParse_after_value
					goto mainparse
				}
			} else {
				goto wantedvalue
			}
		}
	}

handle_URL:

	/* handler: j.URL type=url.URL kind=struct quoted=false*/

	{
		/* Falling back. type=url.URL kind=struct */
		tbuf, err := fs.CaptureField(tok)
		if err != nil {
			return fs.WrapErr(err)
		}

		err = json.Unmarshal(tbuf, &j.URL)
		if err != nil {
			return fs.WrapErr(err)
		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

wantedvalue:
	return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
wrongtokenerror:
	return fs.WrapErr(fmt.Errorf("ffjson: wanted token: %v, but got token: %v output=%s", wantedTok, tok, fs.Output.String()))
tokerror:
	if fs.BigError != nil {
		return fs.WrapErr(fs.BigError)
	}
	err = fs.Error.ToError()
	if err != nil {
		return fs.WrapErr(err)
	}
	panic("ffjson-generated: unreachable, please report bug.")
done:

	return nil
}

// MarshalJSON marshal bytes to json - template
func (j *IfconfigResponse) MarshalJSON() ([]byte, error) {
	var buf fflib.Buffer
	if j == nil {
		buf.WriteString("null")
		return buf.Bytes(), nil
	}
	err := j.MarshalJSONBuf(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalJSONBuf marshal buff to json - template
func (j *IfconfigResponse) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if j == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	buf.WriteString(`{"ip":`)
	fflib.WriteJsonString(buf, string(j.IP))
	buf.WriteString(`,"ip_decimal":`)
	/* json.Number */
	err = buf.Encode(j.IPdecimal)
	if err != nil {
		return err
	}
	buf.WriteString(`,"country":`)
	fflib.WriteJsonString(buf, string(j.Country))
	buf.WriteString(`,"country_iso":`)
	fflib.WriteJsonString(buf, string(j.CountryISO))
	buf.WriteString(`,"city":`)
	fflib.WriteJsonString(buf, string(j.City))
	buf.WriteString(`,"hostname":`)
	fflib.WriteJsonString(buf, string(j.Hostname))
	buf.WriteByte('}')
	return nil
}

const (
	ffjtIfconfigResponsebase = iota
	ffjtIfconfigResponsenosuchkey

	ffjtIfconfigResponseIP

	ffjtIfconfigResponseIPdecimal

	ffjtIfconfigResponseCountry

	ffjtIfconfigResponseCountryISO

	ffjtIfconfigResponseCity

	ffjtIfconfigResponseHostname
)

var ffjKeyIfconfigResponseIP = []byte("ip")

var ffjKeyIfconfigResponseIPdecimal = []byte("ip_decimal")

var ffjKeyIfconfigResponseCountry = []byte("country")

var ffjKeyIfconfigResponseCountryISO = []byte("country_iso")

var ffjKeyIfconfigResponseCity = []byte("city")

var ffjKeyIfconfigResponseHostname = []byte("hostname")

// UnmarshalJSON umarshall json - template of ffjson
func (j *IfconfigResponse) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return j.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

// UnmarshalJSONFFLexer fast json unmarshall - template ffjson
func (j *IfconfigResponse) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error
	currentKey := ffjtIfconfigResponsebase
	_ = currentKey
	tok := fflib.FFTok_init
	wantedTok := fflib.FFTok_init

mainparse:
	for {
		tok = fs.Scan()
		//	println(fmt.Sprintf("debug: tok: %v  state: %v", tok, state))
		if tok == fflib.FFTok_error {
			goto tokerror
		}

		switch state {

		case fflib.FFParse_map_start:
			if tok != fflib.FFTok_left_bracket {
				wantedTok = fflib.FFTok_left_bracket
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_key
			continue

		case fflib.FFParse_after_value:
			if tok == fflib.FFTok_comma {
				state = fflib.FFParse_want_key
			} else if tok == fflib.FFTok_right_bracket {
				goto done
			} else {
				wantedTok = fflib.FFTok_comma
				goto wrongtokenerror
			}

		case fflib.FFParse_want_key:
			// json {} ended. goto exit. woo.
			if tok == fflib.FFTok_right_bracket {
				goto done
			}
			if tok != fflib.FFTok_string {
				wantedTok = fflib.FFTok_string
				goto wrongtokenerror
			}

			kn := fs.Output.Bytes()
			if len(kn) <= 0 {
				// "" case. hrm.
				currentKey = ffjtIfconfigResponsenosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'c':

					if bytes.Equal(ffjKeyIfconfigResponseCountry, kn) {
						currentKey = ffjtIfconfigResponseCountry
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyIfconfigResponseCountryISO, kn) {
						currentKey = ffjtIfconfigResponseCountryISO
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyIfconfigResponseCity, kn) {
						currentKey = ffjtIfconfigResponseCity
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'h':

					if bytes.Equal(ffjKeyIfconfigResponseHostname, kn) {
						currentKey = ffjtIfconfigResponseHostname
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'i':

					if bytes.Equal(ffjKeyIfconfigResponseIP, kn) {
						currentKey = ffjtIfconfigResponseIP
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffjKeyIfconfigResponseIPdecimal, kn) {
						currentKey = ffjtIfconfigResponseIPdecimal
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.EqualFoldRight(ffjKeyIfconfigResponseHostname, kn) {
					currentKey = ffjtIfconfigResponseHostname
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyIfconfigResponseCity, kn) {
					currentKey = ffjtIfconfigResponseCity
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffjKeyIfconfigResponseCountryISO, kn) {
					currentKey = ffjtIfconfigResponseCountryISO
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyIfconfigResponseCountry, kn) {
					currentKey = ffjtIfconfigResponseCountry
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.AsciiEqualFold(ffjKeyIfconfigResponseIPdecimal, kn) {
					currentKey = ffjtIfconfigResponseIPdecimal
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyIfconfigResponseIP, kn) {
					currentKey = ffjtIfconfigResponseIP
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffjtIfconfigResponsenosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			}

		case fflib.FFParse_want_colon:
			if tok != fflib.FFTok_colon {
				wantedTok = fflib.FFTok_colon
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_value
			continue
		case fflib.FFParse_want_value:

			if tok == fflib.FFTok_left_brace || tok == fflib.FFTok_left_bracket || tok == fflib.FFTok_integer || tok == fflib.FFTok_double || tok == fflib.FFTok_string || tok == fflib.FFTok_bool || tok == fflib.FFTok_null {
				switch currentKey {

				case ffjtIfconfigResponseIP:
					goto handle_IP

				case ffjtIfconfigResponseIPdecimal:
					goto handle_IPdecimal

				case ffjtIfconfigResponseCountry:
					goto handle_Country

				case ffjtIfconfigResponseCountryISO:
					goto handle_CountryISO

				case ffjtIfconfigResponseCity:
					goto handle_City

				case ffjtIfconfigResponseHostname:
					goto handle_Hostname

				case ffjtIfconfigResponsenosuchkey:
					err = fs.SkipField(tok)
					if err != nil {
						return fs.WrapErr(err)
					}
					state = fflib.FFParse_after_value
					goto mainparse
				}
			} else {
				goto wantedvalue
			}
		}
	}

handle_IP:

	/* handler: j.IP type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.IP = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_IPdecimal:

	/* handler: j.IPdecimal type=json.Number kind=string quoted=false*/

	{
		/* Falling back. type=json.Number kind=string */
		tbuf, err := fs.CaptureField(tok)
		if err != nil {
			return fs.WrapErr(err)
		}

		err = json.Unmarshal(tbuf, &j.IPdecimal)
		if err != nil {
			return fs.WrapErr(err)
		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Country:

	/* handler: j.Country type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Country = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_CountryISO:

	/* handler: j.CountryISO type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.CountryISO = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_City:

	/* handler: j.City type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.City = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Hostname:

	/* handler: j.Hostname type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Hostname = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

wantedvalue:
	return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
wrongtokenerror:
	return fs.WrapErr(fmt.Errorf("ffjson: wanted token: %v, but got token: %v output=%s", wantedTok, tok, fs.Output.String()))
tokerror:
	if fs.BigError != nil {
		return fs.WrapErr(fs.BigError)
	}
	err = fs.Error.ToError()
	if err != nil {
		return fs.WrapErr(err)
	}
	panic("ffjson-generated: unreachable, please report bug.")
done:

	return nil
}