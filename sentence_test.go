package nmea

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentences(t *testing.T) {
	var sentencetests = []struct {
		name     string
		raw      string
		datatype string
		talkerid string
		prefix   string
		err      string
		sent     BaseSentence
	}{
		{
			name:     "checksum ok",
			raw:      "$GPFOO,1,2,3.3,x,y,zz,*51",
			datatype: "FOO",
			talkerid: "GP",
			prefix:   "GPFOO",
			sent: BaseSentence{
				Talker:   "GP",
				Type:     "FOO",
				Fields:   []string{"1", "2", "3.3", "x", "y", "zz", ""},
				Checksum: "51",
				Raw:      "$GPFOO,1,2,3.3,x,y,zz,*51",
			},
		},
		{
			name:     "trim leading and trailing spaces",
			raw:      "   $GPFOO,1,2,3.3,x,y,zz,*51   ",
			datatype: "FOO",
			talkerid: "GP",
			prefix:   "GPFOO",
			sent: BaseSentence{
				Talker:   "GP",
				Type:     "FOO",
				Fields:   []string{"1", "2", "3.3", "x", "y", "zz", ""},
				Checksum: "51",
				Raw:      "$GPFOO,1,2,3.3,x,y,zz,*51",
			},
		},
		{
			name:     "good parsing",
			raw:      "$GPRMC,235236,A,3925.9479,N,11945.9211,W,44.7,153.6,250905,15.2,E,A*0C",
			datatype: "RMC",
			talkerid: "GP",
			prefix:   "GPRMC",
			sent: BaseSentence{
				Talker:   "GP",
				Type:     "RMC",
				Fields:   []string{"235236", "A", "3925.9479", "N", "11945.9211", "W", "44.7", "153.6", "250905", "15.2", "E", "A"},
				Checksum: "0C",
				Raw:      "$GPRMC,235236,A,3925.9479,N,11945.9211,W,44.7,153.6,250905,15.2,E,A*0C",
			},
		},
		{
			name:     "valid NMEA 4.10 TAG Block",
			raw:      "\\s:Satelite_1,c:1553390539*62\\!AIVDM,1,1,,A,13M@ah0025QdPDTCOl`K6`nV00Sv,0*52",
			datatype: "VDM",
			talkerid: "AI",
			prefix:   "AIVDM",
			sent: BaseSentence{
				Talker:   "AI",
				Type:     "VDM",
				Fields:   []string{"1", "1", "", "A", "13M@ah0025QdPDTCOl`K6`nV00Sv", "0"},
				Checksum: "52",
				Raw:      "!AIVDM,1,1,,A,13M@ah0025QdPDTCOl`K6`nV00Sv,0*52",
				TagBlock: TagBlock{
					Time:   1553390539,
					Source: "Satelite_1",
				},
			},
		},
		{
			name: "checksum bad",
			raw:  "$GPFOO,1,2,3.4,x,y,zz,*51",
			err:  "nmea: sentence checksum mismatch [56 != 51]",
		},
		{
			name: "bad start character",
			raw:  "%GPFOO,1,2,3,x,y,z*1A",
			err:  "nmea: sentence does not start with a '$' or '!'",
		},
		{
			name: "too short prefix",
			raw:  "$XXXX,1,2,3,x,y,z*4B",
			err:  "nmea: sentence address too short: 'XXXX'",
		},
		{
			name: "bad checksum delimiter",
			raw:  "$GPFOO,1,2,3,x,y,z",
			err:  "nmea: sentence does not contain checksum separator",
		},
		{
			name: "no start delimiter",
			raw:  "abc$GPRMC,235236,A,3925.9479,N,11945.9211,W,44.7,153.6,250905,15.2,E,A*0C",
			err:  "nmea: sentence does not start with a '$' or '!'",
		},
		{
			name: "no contain delimiter",
			raw:  "GPRMC,235236,A,3925.9479,N,11945.9211,W,44.7,153.6,250905,15.2,E,A*0C",
			err:  "nmea: sentence does not start with a '$' or '!'",
		},
		{
			name: "another bad checksum",
			raw:  "$GPRMC,235236,A,3925.9479,N,11945.9211,W,44.7,153.6,250905,15.2,E,A*0A",
			err:  "nmea: sentence checksum mismatch [0C != 0A]",
		},
		{
			name: "missing TAG Block start delimiter",
			raw:  "s:Satelite_1,c:1553390539*62\\!AIVDM,1,1,,A,13M@ah0025QdPDTCOl`K6`nV00Sv,0*52",
			err:  "nmea: sentence does not start with a '$' or '!'",
		},
		{
			name: "missing TAG Block end delimiter",
			raw:  "\\s:Satelite_1,c:1553390539*62!AIVDM,1,1,,A,13M@ah0025QdPDTCOl`K6`nV00Sv,0*52",
			err:  "nmea: sentence tag block is missing '\\' at the end",
		},
		{
			name: "invalid TAG Block contents",
			raw:  "\\\\!AIVDM,1,1,,A,13M@ah0025QdPDTCOl`K6`nV00Sv,0*52",
			err:  "nmea: tagblock does not contain checksum separator",
		},
		{
			name: "invalid TAG Block contents",
			raw:  "\\\\!AIVDM,1,1,,A,13M@ah0025QdPDTCOl`K6`nV00Sv,0*52",
			err:  "nmea: tagblock does not contain checksum separator",
		},
	}

	for _, tt := range sentencetests {
		t.Run(tt.name, func(t *testing.T) {
			sent, err := defaultSentenceParser.ParseBaseSentence(tt.raw)
			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.sent, sent)
				assert.Equal(t, tt.sent.Raw, sent.String())
				assert.Equal(t, tt.datatype, sent.DataType())
				assert.Equal(t, tt.talkerid, sent.TalkerID())
				assert.Equal(t, tt.prefix, sent.Prefix())
			}
		})
	}
}

func TestDefaultParseAddress(t *testing.T) {
	var prefixtests = []struct {
		name          string
		prefix        string
		talker        string
		typ           string
		expectedError string
	}{
		{
			name:   "normal prefix",
			prefix: "GPRMC",
			talker: "GP",
			typ:    "RMC",
		},
		{
			name:          "too short, missing type",
			prefix:        "GP",
			talker:        "",
			typ:           "",
			expectedError: `nmea: sentence address too short: 'GP'`,
		},
		{
			name:          "too short, one character",
			prefix:        "X",
			talker:        "",
			typ:           "",
			expectedError: `nmea: sentence address too short: 'X'`,
		},
		{
			name:          "too short",
			prefix:        "GPRM",
			talker:        "",
			typ:           "",
			expectedError: `nmea: sentence address too short: 'GPRM'`,
		},
		{
			name:   "proprietary talker",
			prefix: "PGRME",
			talker: "P",
			typ:    "GRME",
		},
		{
			name:   "short proprietary talker",
			prefix: "PX",
			talker: "P",
			typ:    "X",
		},
		{
			name:   "query",
			prefix: "CCGPQ",
			talker: "CC",
			typ:    "Q",
		},
	}

	for _, tt := range prefixtests {
		t.Run(tt.name, func(t *testing.T) {
			talker, typ, err := DefaultParseAddress(tt.prefix)
			assert.Equal(t, tt.talker, talker)
			assert.Equal(t, tt.typ, typ)
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

var parsetests = []struct {
	name string
	raw  string
	err  error
	msg  interface{}
}{
	{
		name: "bad sentence",
		raw:  "SDFSD,2340dfmswd",
		err:  errors.New("nmea: sentence does not start with a '$' or '!'"),
	},
	{
		name: "bad sentence type",
		raw:  "$INVALID,123,123,*7D",
		err:  &NotSupportedError{Prefix: "INVALID"},
	},
	{
		name: "bad encapsulated sentence type",
		raw:  "!INVALID,1,2,*7E",
		err:  &NotSupportedError{Prefix: "INVALID"},
	},
}

func TestParse(t *testing.T) {
	for _, tt := range parsetests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := Parse(tt.raw)
			if tt.err != nil {
				assert.Equal(t, err, tt.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.msg, m)
			}
		})
	}
}

func TestSentenceParser_Parse(t *testing.T) {
	var testCases = []struct {
		name          string
		givenParser   SentenceParserConfig
		whenInput     string
		expected      Sentence
		expectedError string
	}{
		{
			name:      "ok, parse parametric sentence",
			whenInput: "$HEROT,-11.23,A*07",
			expected: ROT{
				BaseSentence: BaseSentence{
					Talker:   "HE",
					Type:     "ROT",
					Fields:   []string{"-11.23", "A"},
					Checksum: "07",
					Raw:      "$HEROT,-11.23,A*07",
					TagBlock: TagBlock{},
				},
				RateOfTurn: -11.23,
				Valid:      true,
			},
		},
		{
			name:      "ok, parse encapsulated sentence",
			whenInput: "!AIVDM,1,1,,A,13aGt0PP0jPN@9fMPKVDJgwfR>`<,0*55",
			expected: VDMVDO{
				BaseSentence: BaseSentence{
					Talker:   "AI",
					Type:     "VDM",
					Fields:   []string{"1", "1", "", "A", "13aGt0PP0jPN@9fMPKVDJgwfR>`<", "0"},
					Checksum: "55",
					Raw:      "!AIVDM,1,1,,A,13aGt0PP0jPN@9fMPKVDJgwfR>`<,0*55",
					TagBlock: TagBlock{},
				},
				NumFragments:   1,
				FragmentNumber: 1,
				MessageID:      0,
				Channel:        "A",
				Payload: []byte{
					0, 0, 0, 0, 0, 1, 0, 0, 0, 0, // 10
					1, 1, 1, 0, 1, 0, 0, 1, 0, 1,
					0, 1, 1, 1, 1, 1, 1, 1, 0, 0,
					0, 0, 0, 0, 0, 0, 1, 0, 0, 0,
					0, 0, 1, 0, 0, 0, 0, 0, 0, 0, // 50
					0, 0, 0, 0, 1, 1, 0, 0, 1, 0,
					1, 0, 0, 0, 0, 0, 0, 1, 1, 1,
					1, 0, 0, 1, 0, 0, 0, 0, 0, 0,
					1, 0, 0, 1, 1, 0, 1, 1, 1, 0,
					0, 1, 1, 1, 0, 1, 1, 0, 0, 0, // 100
					0, 0, 0, 1, 1, 0, 1, 1, 1, 0,
					0, 1, 1, 0, 0, 1, 0, 1, 0, 0,
					0, 1, 1, 0, 1, 0, 1, 0, 1, 1,
					1, 1, 1, 1, 1, 1, 1, 1, 1, 0,
					1, 1, 1, 0, 1, 0, 0, 0, 1, 0, // 150
					0, 0, 1, 1, 1, 0, 1, 0, 1, 0,
					0, 0, 0, 0, 1, 1, 0, 0,
				},
			},
		},
		{
			name:      "ok, parse query sentence",
			whenInput: "$CCGPQ,GGA*2B",
			expected: Query{
				BaseSentence: BaseSentence{
					Talker:   "CC",
					Type:     "Q",
					Fields:   []string{"GGA"},
					Checksum: "2B",
					Raw:      "$CCGPQ,GGA*2B",
					TagBlock: TagBlock{Time: 0, RelativeTime: 0, Destination: "", Grouping: "", LineCount: 0, Source: "", Text: ""},
				},
				DestinationTalkerID: "GP",
				RequestedSentence:   "GGA",
			},
		},
		{
			name: "ok, parse custom sentence",
			givenParser: SentenceParserConfig{
				CustomParsers: map[string]ParserFunc{
					"YYY": func(s BaseSentence) (Sentence, error) {
						p := NewParser(s)
						return TestZZZ{
							BaseSentence: s,
							NumberValue:  int(p.Int64(0, "number")),
							StringValue:  p.String(1, "str"),
						}, p.Err()

					},
				},
			},
			whenInput: "$AAYYY,20,one,*13",
			expected: TestZZZ{
				BaseSentence: BaseSentence{
					Talker:   "AA",
					Type:     "YYY",
					Fields:   []string{"20", "one", ""},
					Checksum: "13",
					Raw:      "$AAYYY,20,one,*13",
				},
				NumberValue: 20,
				StringValue: "one",
			},
		},
		{
			name:          "nok, sentence prefix too short",
			whenInput:     "$XXXX,20,one,*4A",
			expected:      nil,
			expectedError: "nmea: sentence address too short: 'XXXX'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewSentenceParser()
			if parser != nil {
				parser = NewSentenceParserWithConfig(tc.givenParser)
			}

			result, err := parser.Parse(tc.whenInput)

			assert.Equal(t, tc.expected, result)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSentenceParser_OnTagBlock(t *testing.T) {
	tbCalled := false
	p := NewSentenceParserWithConfig(SentenceParserConfig{OnTagBlock: func(tb TagBlock) {
		tbCalled = true
	}})

	result, err := p.Parse(`\g:1-3-1234,s:r3669961,c:1120959341*0c\`)

	assert.EqualError(t, err, `nmea: sentence does not start with a '$' or '!'`)
	assert.True(t, tbCalled)
	assert.Nil(t, result)
}

func TestSentenceParser_CheckCRC(t *testing.T) {
	var testCases = []struct {
		name          string
		givenCheckCRC func(t *testing.T, sentence BaseSentence, fieldsRaw string) error
		whenInput     string
		expectError   string
		expectCalled  bool
	}{
		{
			name:      "ok, custom CRC check allows invalid CRC",
			whenInput: `$HEROT,-11.23,A*FF`,
			givenCheckCRC: func(t *testing.T, sentence BaseSentence, rawFields string) error {
				assert.Equal(t, "HEROT,-11.23,A", rawFields)
				assert.Equal(t, "FF", sentence.Checksum)
				assert.Equal(t, "ROT", sentence.Type)

				return nil
			},
			expectCalled: true,
		},
		{
			name:      "ok, custom CRC check allows no CRC",
			whenInput: `$HEROT,-11.23,A`,
			givenCheckCRC: func(t *testing.T, sentence BaseSentence, rawFields string) error {
				assert.Equal(t, "HEROT,-11.23,A", rawFields)
				assert.Equal(t, "", sentence.Checksum)
				assert.Equal(t, "ROT", sentence.Type)

				return nil
			},
			expectCalled: true,
		},
		{
			name:      "nok, custom CRC check returns an error",
			whenInput: `$HEROT,-11.23,A`,
			givenCheckCRC: func(t *testing.T, sentence BaseSentence, rawFields string) error {
				assert.Equal(t, "HEROT,-11.23,A", rawFields)
				assert.Equal(t, "", sentence.Checksum)
				assert.Equal(t, "ROT", sentence.Type)

				return errors.New("invalid CRC")
			},
			expectCalled: true,
			expectError:  "invalid CRC",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			p := NewSentenceParserWithConfig(
				SentenceParserConfig{
					CheckCRC: func(sentence BaseSentence, rawFields string) error {
						called = true
						return tc.givenCheckCRC(t, sentence, rawFields)
					},
				},
			)

			_, err := p.Parse(tc.whenInput)

			assert.Equal(t, tc.expectCalled, called)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
