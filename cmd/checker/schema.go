package main

import (
	"fmt"
	"io/ioutil"

	"github.com/cdnjs/tools/util"
	"github.com/xeipuuv/gojsonschema"
)

const (
	validateAgainst = `
{
  "name": "a-happy-tyler",
  "description": "Tyler is happy. Be like Tyler.",
  "keywords": [
    "tyler",
    "happy"
  ],
  "author": {
    "name": "Tyler Caslin",
    "email": "tylercaslin47@gmail.com"
  },
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "git://github.com/tc80/a-happy-tyler.git"
  },
  "filename": "happy.js",
  "autoupdate": {
    "source": "git",
    "target": "git://github.com/tc80/a-happy-tyler.git",
    "fileMap": [
      {
        "basePath": "src",
        "files": [
          "*"
        ]
      }
    ]
  }
}`
	schema = `
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "properties": {
        "authors": {
            "description": "The attributed author for the library, as defined in the cdnjs package JSON file for this library.",
            "type": "array",
            "minItems": 1,
            "uniqueItems": true,
            "items": {
                "type": "object",
                "properties": {
                    "email": {
                        "type": "string",
                        "minLength": 1
                    },
                    "name": {
                        "type": "string",
                        "minLength": 1
                    },
                    "url": {
                        "type": "string",
                        "minLength": 1
                    }
                },
                "additionalProperties": false
            }
        },
        "autoupdate": {
            "description": "Subscribes the package to an autoupdating service when a new version is released.",
            "type": "object",
            "properties": {
                "fileMap": {
                    "type": "array",
                    "minItems": 1,
                    "uniqueItems": true,
                    "items": {
                        "type": "object",
                        "properties": {
                            "basePath": {
                                "type": "string"
                            },
                            "files": {
                                "type": "array",
                                "minItems": 1,
                                "uniqueItems": true,
                                "items": {
                                    "type": "string",
                                    "minLength": 1
                                }
                            }
                        },
                        "required": [
                            "basePath",
                            "files"
                        ],
                        "additionalProperties": false
                    }
                },
                "source": {
                    "type": "string"
                },
                "target": {
                    "type": "string"
                }
            },
            "required": [
                "fileMap",
                "source",
                "target"
            ],
            "additionalProperties": false
        },
        "description": {
            "description": "The description of the library if it has been provided in the cdnjs package JSON file.",
            "type": "string"
        },
        "filename": {
            "description": "This will be the name of the default file for the library.",
            "type": "string"
        },
        "homepage": {
            "description": "A link to the homepage of the package, if one is defined in the cdnjs package JSON file. Normally, this is either the package repository or the package website.",
            "type": "string"
        },
        "keywords": {
            "description": "An array of keywords provided in the cdnjs package JSON for the library.",
            "type": "array",
            "minItems": 1,
            "uniqueItems": true,
            "items": {
                "type": "string"
            }
        },
        "license": {
            "description": "The license defined for the library on cdnjs, as a string. If the library has a custom license, it may not be shown here.",
            "type": "string"
        },
        "name": {
            "description": "This will be the full name of the library, as stored on cdnjs.",
            "type": "string"
        },
        "repository": {
            "description": "The repository for the library, if known, in standard repository format.",
            "type": "object",
            "properties": {
                "directory": {
                    "type": "string"
                },
                "type": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            },
            "required": [
                "type",
                "url"
            ],
            "additionalProperties": false
        }
    },
    "required": [
        "autoupdate",
        "description",
        "filename",
        "keywords",
        "name"
    ],
    "additionalProperties": false
}`
)

func test() {
	// ensure license is valid spdx
	// add tests -- make sure to account for nondeterminism

	testbytes, err := ioutil.ReadFile("cmd/checker/test.json")
	util.Check(err)

	schemabytes, err := ioutil.ReadFile("cmd/checker/schema.json")
	util.Check(err)

	s, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemabytes))
	// s, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schema))
	util.Check(err)

	res, err := s.Validate(gojsonschema.NewBytesLoader(testbytes))
	// res, err := s.Validate(gojsonschema.NewStringLoader(validateAgainst))
	util.Check(err)

	// convert each error to ci error
	fmt.Println(res.Valid(), res.Errors())

	//

	// input := gojsonschema.NewStringLoader(validateAgainst)

	// res, err := gojsonschema.Validate(s, input)
	// fmt.Println(res, err)
	// s, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schema))
	// util.Check(err)

	// gojsonschema.FormatCheckers.Add()
	// res, err := s.Validate(`something`)
	// util.Check(err)
}
