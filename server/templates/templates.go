package templates

import (
	"html/template"
	"log"
	"net/http"
)

/*
Structs
------------
Templates cannot magically access our go variables. In each template we will have logic which will
see who the user is and return a success or unauthorized message. Templates might have logic like:
		{{if .BAlertUser}}
		<div class="alert">{{.AlertMsg}}</div>
		{{end}
Which is important. So for each template we have a struct which is an entrance for sending data from
go to the html templates.
Why Seperate Structs for each template?
Not all templates will have same data required. Some might not need .SecretMsg like restricted page needs,
so seperations of concerns is a best practice here.

So in summary:
This struct represents all the dynamic data that the login template might need.
Templates cannot access local variables from handlers.
Templates can only access data you pass into them.
So if you want your HTML to show:
	an alert or not
	what the alert message is
…you need a data container. That container is this struct.
*/

type LoginPage struct {
	BAlertUser bool
	AlertMsg   string
}

type RegisterPage struct {
	BAlertUser bool
	AlertMsg   string
}

type RestrictedPage struct {
	CsrfSecret    string
	SecretMessage string
}

/*
Here we are parsing all templates at the boot of the server itself. template.ParseFiles() will pass all three
templates and put it in a global registry called templates as *template.Template. This will load everything from disk
and the program can now execute the template by simply calling templates.ExecuteTemplate(w, filename, p)
*/
var templates = template.Must(template.ParseFiles("./server/templates/templatefiles/login.tmpl", "./server/templates/templatefiles/register.tmpl", "./server/templates/templatefiles/restricted.tmpl"))

// This is a render template function which will take a template, its data struct and return an html response of the template
func RenderTemplates(w http.ResponseWriter, tmpl string, p interface{}) {
	//	This will write the template to the http response w, with template we want like login, and its struct data
	err := templates.ExecuteTemplate(w, tmpl+".tmpl", p)
	if err != nil {
		log.Printf("Template error here: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
