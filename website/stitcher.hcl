hostname = "stitcherd.vhodges.dev" 

route "/news/" {
  content {
    source = "website/public/news/index.html"
    
    replacement "#news" {
      content {
        template = "website/message_list.tmpl"
        source = "https://lists.sr.ht/~vhodges/stitcherd-announce/"
        cache = "/news/list/template/"
        ttl = "30m"
      }
    }

    cache = "/news/"
    ttl = "10m"
  }
}

route "/news/{messageId}" {
  content {
    source = "website/public/message/index.html"
    
    replacement ".content" {
      content {
        template = "website/message.tmpl"
        source = "https://lists.sr.ht/~vhodges/stitcherd-announce/{{messageId}}"
        cache = "/news/message/template/{{messageId}}"
        ttl = "30m"
      }
    }

    cache = "/news/{{messageId}}/"
    ttl = "10m"
  }
}

route "/" {
  content {
    source = "website/public/index.html"
    
    replacement "#homepage" {
      content {
        source = "https://github.com/vhodges/stitcherd/blob/main/README.md"
        select = "article:first-of-type"
        cache = "github/homepage"
        ttl = "30m"
      }
    }

    cache = "homepage"
    ttl = "10m"
  } 
}

route "/" {
  static {
    directory = "website/public"
  }
}

