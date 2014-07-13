package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

const lenPath = len("/view/")

var templates = make(map[string]*template.Template)

func init() {
	log.Println("init template cache")
	for _, tmpl := range []string{"edit", "view"} {
		t := template.Must(template.ParseFiles(tmpl + ".html"))
		templates[tmpl] = t
	}
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
	//fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[lenPath:]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
	//	fmt.Fprintf(w, "<h1>Editing %s</h1>"+
	//		"<form> <action = \"/save/%s\" method=\"POST\">"+
	//		"<textarea name=\"body\">%s</textarea><br>"+
	//		"</form>",
	//		p.Title, p.Title, p.Body)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err = p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates[tmpl].Execute(w, p)
	//	t, err := template.ParseFiles(tmpl + ".html")
	//	if err != nil {
	//		log.Fatal(err)
	//		http.Error(w, err.Error(), http.StatusInternalServerError)
	//		return
	//	}
	//  err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getTitle(w http.ResponseWriter, r *http.Request) (title string, err error) {
	title = r.URL.Path[lenPath:]
	titleValidator := regexp.MustCompile("^[a-zA-Z0-9]+$")
	if !titleValidator.MatchString(title) {
		http.NotFound(w, r)
		err = errors.New("invalid page title")
	}
	return
}

func test() {
	p1 := &Page{Title: "TestPage", Body: []byte("This is a sample Page2.\n")}
	p1.save()
	p2, err := loadPage("TestPage")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(p2.Body))
}

func main() {
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	log.Println("Listen... : port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe :", err)
	}
}
