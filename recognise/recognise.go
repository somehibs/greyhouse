package recognise

// lifted from gococo and adapted from cli to api
import (
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"

	"bufio"
	//"bytes"
	"fmt"
	//"image"
	//"image/draw"
	//"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	//"golang.org/x/image/colornames"

	//"golang.org/x/image/font"
	//"golang.org/x/image/font/basicfont"
	//"golang.org/x/image/math/fixed"

)

// Global labels array
type Recogniser struct {
	labels []string
	graph *tf.Graph
}

func NewRecogniser(dataDir string) Recogniser {
	// Load the labels
	labels := loadLabels(dataDir+"/labels.txt")
	graph := loadInference(dataDir)
	return Recogniser{labels, graph}
}

func loadInference(dataDir string) *tf.Graph {
	// Load a frozen graph to use for queries
	modelpath := filepath.Join(dataDir, "frozen_inference_graph.pb")
	model, err := ioutil.ReadFile(modelpath)
	if err != nil {
		log.Fatal(err)
	}

	// Construct an in-memory graph from the serialized form.
	graph := tf.NewGraph()
	if err := graph.Import(model, ""); err != nil {
		log.Fatal(err)
	}
	return graph
}

func (r Recogniser) RecogniseImage(img []byte) []Object {
	// Create a session for inference over graph.
	session, err := tf.NewSession(r.graph, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// DecodeJpeg uses a scalar String-valued tensor as input.
	tensor, err := makeTensorFromBytes(img)
	if err != nil {
		log.Fatal(err)
	}

	// Get all the input and output operations
	inputop := r.graph.Operation("image_tensor")
	// Output ops
	o1 := r.graph.Operation("detection_boxes")
	o2 := r.graph.Operation("detection_scores")
	o3 := r.graph.Operation("detection_classes")
	o4 := r.graph.Operation("num_detections")

	// Execute COCO Graph
	output, err := session.Run(
		map[tf.Output]*tf.Tensor{
			inputop.Output(0): tensor,
		},
		[]tf.Output{
			o1.Output(0),
			o2.Output(0),
			o3.Output(0),
			o4.Output(0),
		},
		nil)
	if err != nil {
		log.Fatal(err)
	}

	ret := []Object{}
	probabilities := output[1].Value().([][]float32)[0]
	classes := output[2].Value().([][]float32)[0]
	//boxes := output[0].Value().([][][]float32)[0]

	for i, probability := range probabilities {
		if probability < 0.4 {
			log.Print("Ignoring item with <.4 confidence")
			continue
		}
		obj := Object{probability, r.labels[int(classes[i])]}
		ret = append(ret, obj)
	}
	return ret
}

type Object struct {
	probability float32
	class string
	//boundingBox []float32
}

// TENSOR UTILITY FUNCTIONS
func makeTensorFromFile(filename string) (*tf.Tensor, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	//r := bytes.NewReader(b)
	//img, _, err := image.Decode(r)
	//if err != nil {
	//	return nil, err
	//}
	return makeTensorFromBytes(b)
}

func makeTensorFromBytes(jpgImage []byte) (*tf.Tensor, error) {
	// DecodeJpeg uses a scalar String-valued tensor as input.
	tensor, err := tf.NewTensor(string(jpgImage))
	if err != nil {
		return nil, err
	}
	// Creates a tensorflow graph to decode the jpeg image
	graph, input, output, err := decodeJpegGraph()
	if err != nil {
		return nil, err
	}
	// Execute that graph to decode this one image
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	normalized, err := session.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		return nil, err
	}
	return normalized[0], nil
}

func decodeJpegGraph() (graph *tf.Graph, input, output tf.Output, err error) {
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)
	output = op.ExpandDims(s,
		op.DecodeJpeg(s, input, op.DecodeJpegChannels(3)),
		op.Const(s.SubScope("make_batch"), int32(0)))
	graph, err = s.Finalize()
	return graph, input, output, err
}

func loadLabels(labelsFile string) []string {
	file, err := os.Open(labelsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	labels := []string{}
	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("ERROR: failed to read %s: %v", labelsFile, err)
	}
	return labels
}

func (r Recogniser) getLabel(idx int, probabilities []float32, classes []float32) string {
	index := int(classes[idx])
	label := fmt.Sprintf("%s (%2.0f%%)", r.labels[index], probabilities[idx]*100.0)

	return label
}

//func addLabel(img *image.RGBA, x, y, class int, label string) {
//	col := colornames.Map[colornames.Names[class]]
//	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}
//
//	d := &font.Drawer{
//		Dst:  img,
//		Src:  image.NewUniform(colornames.Black),
//		Face: basicfont.Face7x13,
//		Dot:  point,
//	}
//
//	//Rect(img, x, y-13, (x + len(label)*7), y-6, 7, col)
//
//	d.DrawString(label)
//}

