ARCHS := linux-amd64 darwin-amd64 darwin-arm64
SRCS := main.go
NAME := awscredget
# BUCKET may be set on command line to upload

targets := $(foreach arch,$(ARCHS),$(NAME)-$(arch))

.PHONY: all clean upload

# Build current architecture 
$(NAME): $(SRCS)
	go build -o $@ $^

# Build all architectures
all: $(NAME) $(targets)

clean:
	rm -f $(targets) $(NAME)

upload: $(targets)
	@if [ "$(BUCKET)" = "" ] ; then echo "Error: BUCKET must be set to use upload" 1>&2 ; exit 1; fi
	aws s3 cp --acl public-read --recursive --exclude '*' --include '$(NAME)-*' . s3://$(BUCKET)/

$(targets): $(SRCS)
	GOOS=$(word 2,$(subst -, ,$@)) GOARCH=$(word 3,$(subst -, ,$@)) \
		go build -o $@ $^
