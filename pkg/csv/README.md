csv
====
Fast CSV Reader/Writer

Usage
----
### Read from file
```go
import "csv"

type FrameInfo struct {
	ActiveImageHeight  int      `csv:"Active Image Height"`
	ActiveImageLeft    int      `csv:"Active Image Left"`
	ActiveImageTop     int      `csv:"Active Image Top"`
	ActiveImageWidth   int      `csv:"Active Image Width"`
	CameraClipName     string   `csv:"Camera Clip Name"`
	CameraRoll         float32  `csv:"Camera Roll"`
	CameraTilt         float32  `csv:"Camera Tilt"`
	MasterTC           string   `csv:"Master TC"`
	MasterTCFrameCount int      `csv:"Master TC Frame Count"`
	SensorFPS          float32  `csv:"Sensor FPS"`
}

type FrameSequence []*FrameInfo

func ReadFile(path string) (FrameSequence, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
    	}
	seq := make(FrameSequence, 0)
	if err := csv.Unmarshal(b, &seq); err != nil {
		return nil, err
	}
	return seq, nil
}
```

### Stop parser on unknown field
```go
func ReadFileUnknown(path string) (FrameSequence, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	dec := csv.NewDecoder(f).SkipUnknown(false)
	c := make(FrameSequence, 0)
	if err := dec.Decode(&c); err != nil {
		return nil, err
    	}
	return c, nil
}
```

### Parse CSV to map[string]string without header
```go
type GenericRecord struct {
	Record map[string]string `csv:,any`
}

type GenericCSV []GenericRecord

func ReadFileIntoMap(path string) (GenericCSV, error) {
	f, err := os.Open(path)
	if err != nil {
	    return nil, err
    	}
    	defer f.Close()
	dec := csv.NewDecoder(f)
	c := make(GenericCSV, 0)
	if err := dec.Decode(&c); err != nil {
	    return nil, err
    	}
	return c, nil
}
```

### Parse CSV from io.Reader stream
```go
func ReadStream(r io.Reader) error {
	dec := csv.NewDecoder(r)

	// read and decode the file header
	line, err := dec.ReadLine()
	if err != nil {
		return err
	}
	if err = dec.DecodeHeader(line); err != nil {
		return err
	}

	// loop until EOF (i.e. dec.ReadLine returns an empty line and nil error);
	// any other error during read will result in a non-nil error
	for {
		// read the next line from stream
		line, err = dec.ReadLine()

		// check for read errors other than EOF
		if err != nil {
			return err
		}

		// check for EOF condition
		if line == "" {
			break
		}

		// decode the record
		v := &FrameInfo{}
		if err = dec.DecodeRecord(v, line); err != nil {
			return err
		}

		// process the record here
		Process(v)
	}
	return nil
}
```