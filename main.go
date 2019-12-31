package main

import (
	"fmt"
	"log"
	"os"

	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/server"
)

var readme string = `# To look up a word, open and read a file with the
# word's name under /words. For example, to read the definition
# of the word 'tree', read the file /words/tree. If the file
# doesn't exist already, it will automatically be created and
# populated with a definition when it is read. Once a definition
# file is opened and read, it will continue to be listed under
# the /words directory. Listing the /words directory will show
# all words looked up so far. 
#
# If you mounted dictfs at the usual location (/mnt/dictfs)
# then you can source this file and then use the following
# functions to query the dictionary. Adapt them as needed if
# you mounted the service somewhere else.

# lookup looks up a word
fn lookup {
	cat /mnt/dictfs/words/^$1
}

# index lists all of the words that the dictionary has looked up so far.
fn index {
	lc /mnt/dictfs/words
}
`

var apiKey string

func lookup(dictFS *fs.FS, word string) (fs.FSNode, error) {
	fmt.Println("Looking up: " + word)
	resp, err := dictQuery(apiKey, word)
	if err != nil {
		return nil, fmt.Errorf("Failed to define " + word)
	}
	if len(resp.defs) == 0 {
		return nil, fmt.Errorf("No definitions for " + word)
	}

	content := resp.responseContent()
	newF := fs.NewStaticFile(dictFS.NewStat(word, "glenda", "glenda", 0444), []byte(content))
	if newF == nil {
		return nil, fmt.Errorf("Can't find /words")
	}
	return newF, nil
}

func WalkFail(dictFS *fs.FS, parent fs.Dir, name string) (fs.FSNode, error) {
	if fs.FullPath(parent) != "/words" {
		return nil, nil
	}
	word := name
	f, err := lookup(dictFS, word)
	if err != nil {
		fmt.Printf("Failed to look up word: %s\n", err)
		return nil, err
	}
	return f, nil
}

func main() {
	apiKey = os.Getenv("DICTIONARY_API_KEY")
	if apiKey == "" {
		log.Fatal("No API key. Define env variable DICTIONARY_API_KEY")
	}
	dictFS := fs.NewFS("glenda", "glenda", 0555,
		fs.WithWalkFailHandler(WalkFail))
	dictFS.Root.AddChild(fs.NewStaticFile(dictFS.NewStat("README", "glenda", "glenda", 0444), []byte(readme)))
	dictFS.Root.AddChild(fs.NewStaticDir(dictFS.NewStat("words", "glenda", "glenda", 0555)))

	server.Serve("0.0.0.0:9999", dictFS)
}
