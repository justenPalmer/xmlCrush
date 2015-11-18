# xmlCrush
Library for flattening and making large XML documents accessible in Go.

## Flattened XML
xmlCrush takes a different approach to XML parsing. It developed out of the need to parse massive, heavily nested and non-standardized XML documents into a usable format within Go. Traditional methods of defining structs that reflect the structure of the XML document were not cutting it because of the flexible nature of the documents. Crawling through the document one node at a time was tedious, produced horrendously ugly code, and was prone to breaking. The solution, xmlCrush, allows all data to be brought in no matter the format - it allows for easy extraction of data based on tags, properties, or even by structural relationship. The code to utilize xmlCrush comes out clean and semantic, and allows for easy storage and manipulation.

In xmlCrush, an XML document is flattened into a content string and array of nodes. The array of nodes define the tags in the XML document along with their corresponding attributes. The content string defines data defined in and around the nodes - it also defines the position of each node, preserving the relationship between tags. This approach, while unconventional, has a number of benefits:

- Full XML document becomes available for manipulation and data extraction with one simple method call
- Flexible and reliable - no matter what the structure of the XML document, xmlCrush can parse it
- All values and properties of the XML document get parsed and are made accessible - it is completely lossless
- Easily store xmlCrush output into any database as an array of nodes and a string
- Extract data from the document with ease by tag, or even tag chains (by relationship of tags)
- Iterate recursively through crushed XML data with a single method call

## Crushing
To crush an XML document from an HTTP endpoint:
```` go
	import(
		"fmt"
		"net/http"
		"xmlCrush"
	)
	resp, err := http.Get("http://pathtoxmldocument.xml")
	defer resp.Body.Close()
	if err != nil {
		//do some error handling
	}
	nodes, content, err := xmlCrush.Crush(resp.Body)
	if err != nil {
		//do some error handling
	}
	//nodes is an array of struct Node defined in XML crush
	fmt.Println("nodes",nodes)

	//content is a string that defines the positions of each node as well as content in and around the nodes
	fmt.Println("content",content)
````
