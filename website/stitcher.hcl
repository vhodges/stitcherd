hostname = "stitcherd.vhodges.dev" 
documentroot = "website/public"

route "/news/" {
  source = "website/public/news/index.html"
  
  replace "#news" {
    with "" {
      template = "website/message_list.tmpl"
      source = "https://lists.sr.ht/~vhodges/stitcherd-announce/"
      cache = "/news/list/template/"
      ttl = "30m"
    }
  }

  cache = "/news/"
  ttl = "10m"
}

route "/news/{messageId}" {
  source = "website/public/message/index.html"
  
  replace ".content" {
    with "" {
      template = "website/message.tmpl"
      source = "https://lists.sr.ht/~vhodges/stitcherd-announce/{{messageId}}"
      cache = "/news/message/template/{{messageId}}"
      ttl = "30m"
    }
  }

  cache = "/news/{{messageId}}/"
  ttl = "10m"
}

route "/community/" {
  source = "website/public/community/index.html"
  
  replace "#message_list" {
    with "" {
      template = "website/message_list.tmpl"
      source = "https://lists.sr.ht/~vhodges/stitcherd-general/"
      cache = "/community/list/template/"
      ttl = "30m"
    }
  }

  cache = "/community/"
  ttl = "10m"
}

// Coming soon... Need to re-work the go template to handle multiple posts 
route "/community/{messageId}" {
  source = "website/public/message/index.html"
  
  replace ".content" {
    with "" {
      template = "website/message.tmpl"
      source = "https://lists.sr.ht/~vhodges/stitcherd-general/{{messageId}}"
      cache = "/community/message/template/{{messageId}}"
      ttl = "30m"
    }
  }

  cache = "/community/{{messageId}}/"
  ttl = "10m"
}

route "/" {
  source = "website/public/index.html"
  
  replace "#homepage" {
    with "" {
      source = "https://github.com/vhodges/stitcherd/blob/main/README.md"
      select = "article:first-of-type"
      cache = "github/homepage"
      ttl = "30m"
    }
  }

  cache = "homepage"
  ttl = "10m"
}
