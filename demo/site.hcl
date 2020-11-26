hostname = "localhost" 

route "/users" {

  content { 
    source = "demo/html/index.html"
    replacement "#replaceme" {
      content {
        template = "demo/template/users.tmpl"
        json = "https://jsonplaceholder.typicode.com/users"
      }
    }
    cache = "/users"
    ttl = "30s"
  }

  maxrate = "10.0"
  burst = "50"

  botmaxrate = "2.0"
  botburst = "5"
}

route "/users/{userid}" {
  content {
    source = "demo/html/index.html"

    replacement "#replaceme" {
      content {
        template = "demo/template/user.tmpl"
        json = "https://jsonplaceholder.typicode.com/users/{{userid}}"
        cache = "JSON:/users/{{userid}}"
        ttl = "5m"
      }
    }

    replacement "#todo_list" {
      content {
        template = "demo/template/todos.tmpl"
        json = "https://jsonplaceholder.typicode.com/todos?userId={{userid}}"
        cache = "JSON:/todos/{{userid}}"
        ttl = "1m"
      }
    }

    cache = "/users/{{userid}}"
    ttl = "10s"
  }
}

// An example of matching /folder/ (ie pretty urls) and inject content in all of them.
route "/{folderPath:.*\\/$}" {
  content  {
    source = "demo/html/{{folderPath}}index.html"

    replacement "#replaceme" {
      content {
        source = "string:<div id='intome'>This is the replacement string (One)</div>"
      }
    }
  }
}

// An example of injecting content into all file paths than end in html
route "/{rest:.*html$}" {
  content {
    source = "demo/html/{{rest}}"
    replacement "#replaceme" {
      content {
        source = "string:<div id='intome'>This is the other replacement string (Two)</div>"
      }
    }
  }
}

// The root of the site
/*
route "/" {
  content  {
    source = "demo/html/index.html"
    replacement "#replaceme" {
      content {
        source = "string:<div id='intome'>This is the other replacement string (Three)</div>"
      }
    }
  }
}
*/

route "/foo" {
  static {
    directory = "demo/foo"
  }
}

route "/" {
  static {
    directory = "demo/html"
  }
}
