// Copyright 2019 Liam White
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package file

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/liamawhite/licenser/pkg/license"
)

func New(license license.Handler) *Mutator {
	return &Mutator{license: license}
}

type Mutator struct {
	license license.Handler
}

func (m *Mutator) AppendLicense(path string, dryRun bool) bool {
	// If we can't detect language skip (return true)
	styled := m.styledLicense(path)
	if styled == nil {
		return true
	}
	contents := getFileContents(path)
	if contents == nil {
		return false
	}
	if !m.license.IsPresent(bytes.NewReader(contents)) {
		newContents := append(styled, contents...)
		if dryRun {
			fmt.Printf("%s\n", newContents)
		} else {
			if err := ioutil.WriteFile(path, newContents, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "error writing license to %v:%v", path, err)
			}
		}
	}
	return true
}

func (m *Mutator) VerifyLicense(path string, _ bool) bool {
	contents := getFileContents(path)
	if contents == nil {
		return false
	}
	// If we can't detect language skip (return true)
	if style := identifyLanguageStyle(path); style == nil {
		return true
	}
	present := m.license.IsPresent(bytes.NewReader(contents))
	if !present {
		fmt.Fprintf(os.Stderr, "license missing from %v\n", path)
	}
	return present
}

// this should probably be cached on a per language basis
func (m *Mutator) styledLicense(path string) []byte {
	style := identifyLanguageStyle(path)
	if style == nil {
		return nil
	}
	buf := bytes.NewBuffer([]byte{})

	// TODO: implement block styling
	if style.isBlock {

	} else {
		scanner := bufio.NewScanner(m.license.Reader())
		for scanner.Scan() {
			buf.WriteString(style.comment)
			if len(scanner.Bytes()) != 0 {
				buf.WriteString(" ")
			}
			buf.Write(scanner.Bytes())
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}
	return buf.Bytes()
}

// This function has the potential to become an unwiedly mess, consider rethinking.
func identifyLanguageStyle(path string) *languageStyle {
	switch filepath.Ext(path) {
	case ".go":
		return commentStyles["golang"]
	case ".py":
		return commentStyles["python"]
	case ".sh":
		return commentStyles["shell"]
	default:
		fmt.Fprintf(os.Stderr, "unable to identify language of %v\n", path)
		return nil
	}
}

func getFileContents(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to open %v", path)
	}

	// This will be an issue with really large files...
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read file %v", path)
	}
	return contents
}