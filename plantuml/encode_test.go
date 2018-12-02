package plantuml

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

var plantUMLJarPath = os.Getenv("PLANTUML_JAR_PATH")

func TestEncode(t *testing.T) {
	// NOTE: this was confirmed to be a valid encoding by loading it using http://www.plantuml.com/plantuml/uml/{encoded_string}
	const expectedEncoding = "UDfqaB5A134GXliynR1zk5R28TVebGftsgZ5a0ooS4x2P7PHnBVmti7k7ijx39cDQx5W4mRc-_xypAOgYWul1uTHiq0tL997Ff-EvzFn-_FhRd_jBph02hPm1KNxs1TaXQuSa4trHVBhsmHDqpzK9Hb99goDpbac-Ckc1RYzsUrXLSTdU0XcnCnAh4_MhHzHQSwtrxoM4s1xrRpLUo2EEL6QtOzVUEt_O24TX65tluM4Ygt4siybbkj4YzytnCaEzWC00F__0G00___n7NB4"
	b, err := ioutil.ReadFile("../testdata/sequence.uml")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}
	if exp, got := expectedEncoding, Encode(b); exp != got {
		t.Errorf("encoding not as expected:\nexp: %s\ngot: %s", exp, got)
	}
}

func TestEncodeAgainstPlantUMLJar(t *testing.T) {
	if plantUMLJarPath == "" {
		t.SkipNow()
	}
	b, err := ioutil.ReadFile("../testdata/sequence.uml")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}
	encodedText := Encode(b)

	out := &bytes.Buffer{}
	cmd := exec.Command("java", "-jar", plantUMLJarPath, "-decodeurl", encodedText)
	cmd.Stdout = out
	err = cmd.Run()
	if err != nil {
		t.Fatalf("failed to run command: %v", err)
	}

	// NOTE: annoyingly, the -decodeurl command seems to print out two @startuml and
	//		 two @enduml tags. This code below attempts to handle any number of start
	//		 and end tags so as to be future proof in case this changes. The end result
	//		 is to have just one @startuml and one @enduml (followed by newline).
	startIndex := bytes.LastIndex(out.Bytes(), []byte("@startuml"))
	endTag := []byte("@enduml")
	endIndex := bytes.Index(out.Bytes(), []byte("@enduml"))

	if startIndex == -1 || endIndex == -1 {
		t.Fatalf("failed to find start and end tags in output:\n%s", string(out.Bytes()))
	}
	endIndex += len(endTag)
	if endIndex < out.Len() {
		// for the newline
		endIndex++
	}
	if exp, got := b, out.Bytes()[startIndex:endIndex]; !reflect.DeepEqual(exp, got) {
		t.Errorf("encoded uml not as expected:\nexp:\n%s\ngot:\n%s", string(exp), string(got))
	}
}
