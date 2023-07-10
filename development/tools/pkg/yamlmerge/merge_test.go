package yamlmerge

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
)

func createTempfiles(noYaml, noOther int) (string, func(string)) {
	dir, fn := createTempDir()
	for i := 0; i < noYaml; i++ {
		tmpfn := filepath.Join(dir, fmt.Sprintf("%s%d%s", "tmpfile", i, ".yaml"))
		if err := ioutil.WriteFile(tmpfn, []byte("tmpcontent"), 0666); err != nil { // no newline at the end of the file will force the merge function to add it
			log.Fatal(err)
		}
	}
	for i := 0; i < noOther; i++ {
		tmpfn := filepath.Join(dir, fmt.Sprintf("%s%d%s", "tmpfile", i, ".other"))
		if err := ioutil.WriteFile(tmpfn, []byte("tmpcontent\n"), 0666); err != nil {
			log.Fatal(err)
		}
	}
	return dir, fn
}

func createStarterYamlTemps() (string, func(string)) {
	dir, fn := createTempDir()
	tmpfn := filepath.Join(dir, fmt.Sprintf("%s%s", "starter", ".yaml"))
	if err := ioutil.WriteFile(tmpfn, []byte("starter\n"), 0666); err != nil {
		log.Fatal(err)
	}
	return dir, fn
}

func createTempDir() (string, func(string)) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatalf("Coulnd't create tempdir")
	}
	return dir, func(dir string) {
		os.RemoveAll(dir) // cleanup after ourselves
	}
}

func TestMergeFiles(t *testing.T) {
	type args struct {
		path       string
		extension  string
		target     string
		changeFile bool
	}

	mixedFiles, cleanMixedFn := createTempfiles(2, 2)
	startYamlFiles, cleanStarterYamlFn := createStarterYamlTemps()

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Read starter.yaml, should have content of two files", args: args{path: mixedFiles, extension: ".yaml", target: fmt.Sprintf("%s%s%s", mixedFiles, string(os.PathSeparator), "starter.yaml"), changeFile: true}, want: "tmpcontent\ntmpcontent\n"},
		{name: "StarterYaml should not be modified", args: args{path: startYamlFiles, extension: ".yaml", target: fmt.Sprintf("%s%s%s", startYamlFiles, string(os.PathSeparator), "starter.yaml"), changeFile: true}, want: "starter\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MergeFiles(tt.args.path, tt.args.extension, tt.args.target, tt.args.changeFile)

			got, err := ioutil.ReadFile(tt.args.target)
			if err != nil {
				t.Errorf("File was not created or could not be read: %s", err.Error())
			}
			if !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("MergeFiles() = %v, want %v", string(got), tt.want)
			}
		})
	}

	cleanMixedFn(mixedFiles)
	cleanStarterYamlFn(startYamlFiles)
}

func TestMergeFilesDryRun(t *testing.T) {
	type args struct {
		path       string
		extension  string
		target     string
		changeFile bool
	}

	mixedFiles, cleanMixedFn := createTempfiles(2, 2)
	startYamlFiles, cleanStarterYamlFn := createStarterYamlTemps()

	tests := []struct {
		name string
		args args
	}{
		{name: "Try to read starter.yaml, but file should not be created", args: args{path: mixedFiles, extension: ".yaml", target: fmt.Sprintf("%s%s%s", mixedFiles, string(os.PathSeparator), "starter.yaml"), changeFile: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MergeFiles(tt.args.path, tt.args.extension, tt.args.target, tt.args.changeFile)

			_, err := ioutil.ReadFile(tt.args.target)
			if err == nil {
				t.Errorf("starter.yaml should not exist.")
			}
		})
	}

	cleanMixedFn(mixedFiles)
	cleanStarterYamlFn(startYamlFiles)
}

func TestMergeFilesNoYamlFiles(t *testing.T) {
	type args struct {
		path       string
		extension  string
		target     string
		changeFile bool
	}

	onlyOtherFiles, cleanOtherFn := createTempfiles(0, 2)

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "No file created, test should fail, due to no yaml files in path", args: args{path: onlyOtherFiles, extension: ".yaml", target: fmt.Sprintf("%s%s%s", onlyOtherFiles, string(os.PathSeparator), "starter.yaml"), changeFile: true}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MergeFiles(tt.args.path, tt.args.extension, tt.args.target, tt.args.changeFile)

			got, err := ioutil.ReadFile(tt.args.target)
			if err == nil {
				t.Errorf("File should not exist, there's no yaml files on path.")
			}
			if !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("MergeFiles() = %v, want %v", got, tt.want)
			}
		})
	}

	cleanOtherFn(onlyOtherFiles)
}

func Test_collectFiles(t *testing.T) {
	type args struct {
		path      string
		extension string
	}

	mixedFiles, cleanMixedFn := createTempfiles(2, 2)
	onlyYamlFiles, cleanYamlFn := createTempfiles(2, 0)
	onlyOtherFiles, cleanOtherFn := createTempfiles(0, 2)

	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "Collect only yaml files", args: args{mixedFiles, ".yaml"}, want: []string{fmt.Sprintf("%s%stmpfile0.yaml", mixedFiles, string(os.PathSeparator)), fmt.Sprintf("%s%stmpfile1.yaml", mixedFiles, string(os.PathSeparator))}},
		{name: "Collect only yaml files, when there's no other files", args: args{onlyYamlFiles, ".yaml"}, want: []string{fmt.Sprintf("%s%stmpfile0.yaml", onlyYamlFiles, string(os.PathSeparator)), fmt.Sprintf("%s%stmpfile1.yaml", onlyYamlFiles, string(os.PathSeparator))}},
		{name: "No yaml files on path", args: args{onlyOtherFiles, ".yaml"}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectFiles(tt.args.path, tt.args.extension); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collectFiles() = %v, want %v", got, tt.want)
			}
		})
	}
	cleanMixedFn(mixedFiles)
	cleanYamlFn(onlyYamlFiles)
	cleanOtherFn(onlyOtherFiles)
}

func Test_removeFromArray(t *testing.T) {
	type args struct {
		s []string
		r string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "Remove string successful", args: args{s: []string{"a", "b", "c"}, r: "b"}, want: []string{"a", "c"}},
		{name: "Remove only first find successful", args: args{s: []string{"a", "b", "c", "b"}, r: "b"}, want: []string{"a", "c", "b"}},
		{name: "Remove nothing", args: args{s: []string{"a", "b", "c"}, r: "d"}, want: []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeFromArray(tt.args.s, tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeFromArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
