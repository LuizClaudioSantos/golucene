package document

import (
	"bytes"
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis"
	. "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	"github.com/balzaczyy/golucene/core/index/model"
	"io"
	"log"
	"strconv"
)

// document/Field.java

type Field struct {
	_type  *FieldType  // Field's type
	_name  string      // Field's name
	_data  interface{} // Field's value
	_boost float32     // Field's boost

	/*
		Pre-analyzed tokenStream for indexed fields; this is
		separte from fieldsData because you are allowed to
		have both; eg maybe field has a String value but you
		customize how it's tokenized
	*/
	_tokenStream analysis.TokenStream
}

// Create field with String value
func NewStringField(name, value string, ft *FieldType) *Field {
	assert2(name != "", "name cannot be empty")
	assert2(value != "", "value cannot be empty")
	assert2(ft.stored || ft.indexed,
		"it doesn't make sense to have a field that is neither indexed nor stored")
	return &Field{_type: ft, _name: name, _data: value, _boost: 1}
}

func (f *Field) StringValue() string {
	switch f._data.(type) {
	case string:
		return f._data.(string)
	case int:
		return strconv.Itoa(f._data.(int))
	default:
		log.Println("Unknown type", f._data)
		panic("not implemented yet")
	}
}

func assert2(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}

func (f *Field) ReaderValue() io.Reader {
	if v, ok := f._data.(io.Reader); ok {
		return v
	}
	return nil
}

func (f *Field) Name() string {
	return f._name
}

func (f *Field) Boost() float32 {
	return f._boost
}

func (f *Field) NumericValue() interface{} {
	switch f._data.(type) {
	case int32, int64, float32, float64:
		return f._data
	default:
		return nil
	}
}

func (f *Field) BinaryValue() []byte {
	if v, ok := f._data.([]byte); ok {
		return v
	}
	return nil
}

func (f *Field) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%v<%v:", f._type, f._name)
	if f._data != nil {
		fmt.Fprint(&buf, f._data)
	}
	fmt.Fprint(&buf, ">")
	return buf.String()
}

func (f *Field) FieldType() model.IndexableFieldType {
	return f._type
}

func (f *Field) TokenStream(analyzer analysis.Analyzer, reuse analysis.TokenStream) (ts analysis.TokenStream, err error) {
	if !f.FieldType().Indexed() {
		return nil, nil
	}

	if nt := f.FieldType().(*FieldType).NumericType(); nt != NumericType(0) {
		panic("not implemented yet")
	}

	if !f.FieldType().Tokenized() {
		assert2(f.StringValue() != "", "Non-Tokenized Fields must have a string value")
		if _, ok := reuse.(*StringTokenStream); !ok {
			reuse = newStringTokenStream()
		}
		reuse.(*StringTokenStream).setValue(f.StringValue())
		return reuse, nil
	}

	if f._tokenStream != nil {
		return f._tokenStream, nil
	} else if f.ReaderValue() != nil {
		panic("not implemented yet")
	} else if f.StringValue() != "" {
		return analyzer.TokenStreamForString(f._name, f.StringValue())
	}

	panic(fmt.Sprintf("Field must have either TokenStream, String, Reader, or Number value; got %v", f))
}

type StringTokenStream struct {
	*analysis.TokenStreamImpl
	termAttribute   CharTermAttribute
	offsetAttribute OffsetAttribute
	used            bool
	value           string
}

/*
Creates a new TokenStream that returns a string as single token.

Warning: Does not initialize the value, you must call setValue() afterwards!
*/
func newStringTokenStream() *StringTokenStream {
	ans := &StringTokenStream{TokenStreamImpl: analysis.NewTokenStream()}
	ans.termAttribute = ans.Attributes().Add("CharTermAttribute").(CharTermAttribute)
	ans.offsetAttribute = ans.Attributes().Add("OffsetAttribute").(OffsetAttribute)
	return ans
}

func (ts *StringTokenStream) setValue(value string) {
	ts.value = value
}

func (ts *StringTokenStream) IncrementToken() (bool, error) {
	if ts.used {
		return false, nil
	}
	ts.Attributes().Clear()
	ts.termAttribute.AppendString(ts.value)
	ts.offsetAttribute.SetOffset(0, len(ts.value))
	ts.used = true
	return true, nil
}

/* Specifies whether and how a field should be stored. */
type Store int

/*
Store the original field value in the index. This is useful for short
texts like a document's title which should be displayed with the
results. The value is stored in its original form, i.e. no analyzer
is used before it is stored.
*/
const STORE_YES = Store(1)

/* Do not store the field's value in the index. */
const STORE_NO = Store(2)

// document/TextField.java

var (
	// Indexed, tokenized, not stored
	TEXT_FIELD_TYPE_NOT_STORED = func() *FieldType {
		ft := newFieldType()
		ft.indexed = true
		ft._omitNorms = true
		ft._indexOptions = model.INDEX_OPT_DOCS_ONLY
		ft._tokenized = false
		ft.frozen = true
		return ft
	}()
	// Indexed, tokenized, stored
	TEXT_FIELD_TYPE_STORED = func() *FieldType {
		ft := newFieldType()
		ft.indexed = true
		ft._omitNorms = true
		ft._indexOptions = model.INDEX_OPT_DOCS_ONLY
		ft.stored = true
		ft._tokenized = false
		ft.frozen = true
		return ft
	}()
)

type TextField struct {
	*Field
}

func NewTextField(name, value string, store Store) *TextField {
	return &TextField{NewStringField(name, value, map[Store]*FieldType{
		STORE_YES: TEXT_FIELD_TYPE_STORED,
		STORE_NO:  TEXT_FIELD_TYPE_NOT_STORED,
	}[store])}
}

// document/StoredField.java

// Type for a stored-only field.
var STORED_FIELD_TYPE = func() *FieldType {
	ans := newFieldType()
	ans.stored = true
	return ans
}()

/*
A field whose value is stored so that IndexSearcher.doc()
and IndexReader.document() will return the field and its
value.
*/
type StoredField struct {
	*Field
}

/*
Create a stored-only field with the given binary value.

NOTE: the provided byte[] is not copied so be sure
not to change it until you're done with this field.
*/
// func newStoredField(name string, value []byte) *StoredField {
// 	return &StoredField{newStringField(name, value, STORED_FIELD_TYPE)}
// }
