package compiler

import (
	"sort"
	"sync"
)

var (
	constructorsMu sync.RWMutex
	constructors   = make(map[string]Constructor)
)

// Constructor defines the type for a compiler constructor
type Constructor func() (Compiler, error)

// Register makes a compiler constructor available for the specified language.
// panic if Register is called twice with the same language or if compiler is nil
func Register(language string, cnstr Constructor) {
	constructorsMu.Lock()
	defer constructorsMu.Unlock()
	if len(language) <= 0 {
		panic("compiler Register: language cannot be empty")
	}
	if cnstr == nil {
		panic("compiler Register: constructor cannot be nil")
	}
	if _, dup := constructors[language]; dup {
		panic("compiler Register: cannot be called twice for constructor of " + language)
	}
	constructors[language] = cnstr
}

func unregisterAllCompilers() {
	constructorsMu.Lock()
	defer constructorsMu.Unlock()
	// For tests.
	constructors = make(map[string]Constructor)
}

// Languages returns a sorted list of the language of the registered compiler constructors
func Languages() []string {
	constructorsMu.RLock()
	defer constructorsMu.RUnlock()
	list := make([]string, 0, len(constructors))
	for language := range constructors {
		list = append(list, language)
	}
	sort.Strings(list)
	return list
}

// constructor returns the compiler constructor for a specific language
func constructor(language string) Constructor {
	constructorsMu.Lock()
	defer constructorsMu.Unlock()
	return constructors[language]
}
