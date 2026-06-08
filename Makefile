.PHONY: generate\:key

generate\:key:
	@printf 'APP_KEY='
	@openssl rand -hex 32
