prefix = /usr
share = /usr/share

install:
	install -D gcy $(DESTDIR)$(prefix)/bin/gcy
	install -D autocomplete/completion.bash $(DESTDIR)$(share)/bash-completion/completions/gcy
	install -D autocomplete/completion.zsh $(DESTDIR)$(share)/zsh/site-functions/_gcy

uninstall:
	rm -rf $(DESTDIR)$(prefix)/bin/gcy $(DESTDIR)$(share)/bash-completion/completions/gcy $(DESTDIR)$(share)/zsh/site-functions/_gcy

manpages:
	install -D man/gcy.1 $(DESTDIR)$(share)/man/man1/gcy.1
	install -D man/gcy-kms.5 $(DESTDIR)$(share)/man/man5/gcy-kms.5
	install -D man/gcy-gpg.5 $(DESTDIR)$(share)/man/man5/gcy-gpg.5
	install -D man/gcy-password.5 $(DESTDIR)$(share)/man/man5/gcy-password.5
