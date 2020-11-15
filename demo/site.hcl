hostname = "localhost" 
documentroot = "demo/html"  # Optional, allow static content from this folder

route "/users" {
  source = "demo/html/index.html"
  
  replace "#replaceme" {
    with "" {
      template = "demo/template/users.tmpl"
      json = "https://jsonplaceholder.typicode.com/users"
    }
  }

  cache = "/users"
  ttl = "30s"
}

route "/users/{userid}" {
  source = "demo/html/index.html"
  
  replace "#replaceme" {
    with "" {
      template = "demo/template/user.tmpl"
      json = "https://jsonplaceholder.typicode.com/users/{{userid}}"
      cache = "JSON:/users/{{userid}}"
      ttl = "5m"
    }
  }

  replace "#todo_list" {
    with "" {
      template = "demo/template/todos.tmpl"
      json = "https://jsonplaceholder.typicode.com/todos?userId={{userid}}"
      cache = "JSON:/todos/{{userid}}"
      ttl = "1m"
    }
  }

  cache = "/users/{{userid}}"
  ttl = "10s"
}

// An example of matching /folder/ (ie pretty urls) and inject content in all of them.
route "/{folderPath:.*\\/$}" {

  source = "demo/html/{{folderPath}}index.html"

  replace "#replaceme" {
    with "" {
      source = "string:<div id='replaceme'>This is the other replacement string</div>"
    }
  }
}

// An example of injecting content into all file paths than end in html
route "/{rest:.*html$}" {

  source = "demo/html/{{rest}}"

  replace "#replaceme" {
    with "" {
      source = "string:<div id='replaceme'>This is the other other replacement string</div>"
    }
  }
}

// The root of the site
route "/" {
  source = "demo/html/index.html"
}

