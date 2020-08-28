docker build -t clever -f Docker.swagger. \
  && docker run --rm -it -p 8080:8080 clever
