package aifs

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestOpen(t *testing.T) {

	aifs := New(os.DirFS("testdata"))

	type Case struct {
		in string
		ok bool
	}

	cases := []Case{
		{
			in: `.`,
			ok: true,
		},
		{
			in: `cat1`,
			ok: true,
		},
		{
			in: `bat1`,
			ok: false,
		},
		{
			in: `cat1/cat2`,
			ok: true,
		},
		{
			in: `root.zip`,
			ok: true,
		},
		{
			in: `root.zip/cat1744.jpg`,
			ok: true,
		},
		{
			in: `dirinzip.zip`,
			ok: true,
		},
		{
			in: `dirinzip.zip/z`,
			ok: true,
		},
		{
			in: `dirinzip.zip/z/cat1744.jpg`,
			ok: true,
		},
	}

	for _, c := range cases {
		func() {
			f, err := aifs.Open(c.in)
			if err != nil {
				if c.ok {
					t.Errorf("case: %v should pass: %v", c.in, err)
				}
			} else {
				defer f.Close()
				if !c.ok {
					t.Errorf("case: %v should not pass", c.in)
				}
			}
		}()
	}
}

func TestReadDir(t *testing.T) {

	aifs := New(os.DirFS("testdata"))

	type Case struct {
		in string
		ok bool
	}

	cases := []Case{
		{
			in: `.`,
			ok: true,
		},
		{
			in: `cat1`,
			ok: true,
		},
		{
			in: `bat1`,
			ok: false,
		},
		{
			in: `cat1/cat2`,
			ok: true,
		},
		{
			in: `root.zip`,
			ok: true,
		},
		{
			in: `root.zip/cat1744.jpg`,
			ok: false,
		},
		{
			in: `dirinzip.zip`,
			ok: true,
		},
		{
			in: `dirinzip.zip/z`,
			ok: true,
		},
		{
			in: `dirinzip.zip/z/cat1744.jpg`,
			ok: false,
		},
		{
			in: `にほんご.zip`,
			ok: true,
		},
		{
			in: `にほんご.zip/にほんご`,
			ok: true,
		},
	}

	for _, c := range cases {
		func() {
			_, err := aifs.ReadDir(c.in)
			if err != nil {
				if c.ok {
					t.Errorf("case: %v should pass: %v", c.in, err)
				}
			} else {
				if !c.ok {
					t.Errorf("case: %v should not pass", c.in)
				}
			}
		}()
	}

}

func TestSepFilePath(t *testing.T) {
	type Case struct {
		in   string
		want []string
	}

	cases := []Case{
		{
			in:   `hoge/huga/piyo`,
			want: []string{`hoge`, `huga`, `piyo`},
		},
		{
			in:   `.`,
			want: []string{`.`},
		},
		{
			in:   `zip.zip/`,
			want: []string{`zip.zip`},
		},
	}

	for _, c := range cases {
		got := sepFilePath(filepath.Clean(c.in))

		if !reflect.DeepEqual(c.want, got) {
			t.Errorf("want :%v, got: %v", c.want, got)
		}
	}

}
