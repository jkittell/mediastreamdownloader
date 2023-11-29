package downloader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jkittell/array"
	"github.com/jkittell/mediastreamparser/parser"
	"github.com/jkittell/toolbox"
	"log"
	"os"
	"path"
)

func (s Stream) JSON() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(s)
	return buffer.Bytes(), err
}

type Stream struct {
	Name string `json:"name"`
	File string `json:"file"`
}

func getStreamSegments(segments *array.Array[parser.Segment], streamName string) *array.Array[parser.Segment] {
	streamSegments := array.New[parser.Segment]()
	for i := 0; i < segments.Length(); i++ {
		seg := segments.Lookup(i)
		if seg.StreamName == streamName {
			streamSegments.Push(seg)
		}
	}
	return streamSegments
}

func combineSegmentsToFile(dir, out string, segments *array.Array[parser.Segment]) string {
	files := array.New[string]()
	for i := 0; i < segments.Length(); i++ {
		seg := segments.Lookup(i)
		name := path.Join(dir, seg.SegmentName)
		size, err := toolbox.DownloadFile(name, seg.SegmentURL, nil)
		if err != nil {
			log.Println(err)
		}
		if size > 0 {
			files.Push(name)
		} else {
			log.Println("file size is zero bytes")
		}
	}

	var data []byte
	for i := 0; i < files.Length(); i++ {
		f := files.Lookup(i)
		b, err := os.ReadFile(f)
		if err != nil {
			log.Println(err)
		}
		data = append(data, b...)
		err = toolbox.DeleteFile(f)
		if err != nil {
			log.Println(err)
		}
	}
	err := os.WriteFile(out, data, 0644)
	if err != nil {
		log.Println(err)
	}

	return out
}

// Contains the value in the array
func contains(arr *array.Array[string], name string) bool {
	for i := 0; i < arr.Length(); i++ {
		j := arr.Lookup(i)
		if name == j {
			return true
		}
	}
	return false
}

func Run(dir, playlistURL string) *array.Array[Stream] {
	results := array.New[Stream]()

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Println(err)
		return results
	}

	streams := array.New[string]()
	segments, err := parser.GetSegments(playlistURL)
	if err != nil {
		log.Println("unable to get segments")
	}

	for i := 0; i < segments.Length(); i++ {
		seg := segments.Lookup(i)
		if contains(streams, seg.StreamName) {
			continue
		} else {
			streams.Push(seg.StreamName)
		}
	}

	for i := 0; i < streams.Length(); i++ {
		strName := streams.Lookup(i)
		strSegments := getStreamSegments(segments, strName)
		if strSegments.Length() > 0 {
			name := path.Join(dir, fmt.Sprintf("%s.mp4", strName))
			combineSegmentsToFile(dir, name, strSegments)
			if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
				log.Printf("%s does not exist", name)
			}
			res := Stream{
				Name: strName,
				File: name,
			}
			results.Push(res)
		} else {
			log.Printf("no stream segments found for '%s'\n", strName)
		}
	}

	if err != nil {
		log.Println(err)
		return results
	}
	return results
}
