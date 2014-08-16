package axmlParser

import (
	"fmt"
	"io"
)

type Manifest struct {
	Attrs map[string][]*Attribute
}

type PlainListener struct {
	Manifest Manifest
}

func (listener *PlainListener) BuildXml(writer io.Writer) error {
	return nil
}

func (listener *PlainListener) StartDocument() {
	listener.Manifest.Attrs = make(map[string][]*Attribute)
}

/**
 * Receive notification of the end of a document.
 */
func (listener *PlainListener) EndDocument() {
}

/**
 * Begin the scope of a prefix-URI Namespace mapping.
 *
 * @param prefix
 *            the Namespace prefix being declared. An empty string is used
 *            for the default element namespace, which has no prefix.
 * @param uri
 *            the Namespace URI the prefix is mapped to
 */
func (listener *PlainListener) StartPrefixMapping(prefix, uri string) {
}

/**
 * End the scope of a prefix-URI mapping.
 *
 * @param prefix
 *            the prefix that was being mapped. This is the empty string
 *            when a default mapping scope ends.
 * @param uri
 *            the Namespace URI the prefix is mapped to
 */
func (listener *PlainListener) EndPrefixMapping(prefix, uri string) {}

/**
 * Receive notification of the beginning of an element.
 *
 * @param uri
 *            the Namespace URI, or the empty string if the element has no
 *            Namespace URI or if Namespace processing is not being
 *            performed
 * @param localName
 *            the local name (without prefix), or the empty string if
 *            Namespace processing is not being performed
 * @param qName
 *            the qualified name (with prefix), or the empty string if
 *            qualified names are not available
 * @param atts
 *            the attributes attached to the element. If there are no
 *            attributes, it shall be an empty Attributes object. The value
 *            of this object after startElement returns is undefined
 */
func (listener *PlainListener) StartElement(uri, localName, qName string,
	attrs []*Attribute) {
	for _, attr := range attrs {
		if _, ok := listener.Manifest.Attrs[localName]; !ok {
			listener.Manifest.Attrs[localName] = make([]*Attribute, 0)
		}
		listener.Manifest.Attrs[localName] = append(listener.Manifest.Attrs[localName], attr)
		fmt.Println(localName, attr)
	}
}

/**
 * Receive notification of the end of an element.
 *
 * @param uri
 *            the Namespace URI, or the empty string if the element has no
 *            Namespace URI or if Namespace processing is not being
 *            performed
 * @param localName
 *            the local name (without prefix), or the empty string if
 *            Namespace processing is not being performed
 * @param qName
 *            the qualified XML name (with prefix), or the empty string if
 *            qualified names are not available
 */
func (listener *PlainListener) EndElement(uri, localName, qName string) {}

/**
 * Receive notification of text.
 *
 * @param data
 *            the text data
 */
func (listener *PlainListener) Text(data string) {}

/**
 * Receive notification of character data (in a <![CDATA[ ]]> block).
 *
 * @param data
 *            the text data
 */
func (listener *PlainListener) CharacterData(data string) {}

/**
 * Receive notification of a processing instruction.
 *
 * @param target
 *            the processing instruction target
 * @param data
 *            the processing instruction data, or null if none was supplied.
 *            The data does not include any whitespace separating it from
 *            the target
 * @throws org.xml.sax.SAXException
 *             any SAX exception, possibly wrapping another exception
 */
func (listener *PlainListener) ProcessingInstruction(target, data string) {

}
