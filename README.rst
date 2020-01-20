poŝtisto
========

|license| |release| |gitter| |build| |godocs| |gomod| |codecov| |goreport| |codebeat|

.. contents::
    :backlinks: none
    :local:


General
-------

poŝtisto is the successor of Tabellarius, an IMAP client that sorts your mailboxes based on an extensive configuration language.
Unlike other mail filters it uses the `same IMAP connection accross multiple IMAP commands <https://github.com/lefcha/imapfilter>`_ and `simple markup language <http://www.rfcreader.com/#rfc5228>`_ instead of a complex scripting language though it isn't that feature-rich as the well-known Sieve standard.

It became necessary because of missing features in the ManageSieve protocol and service providers that don't even provide a ManageSieve service or any *more advanced* filter techniques.

What it actually does is to parse your YAML config files, sets up an IMAP connection pool to one or more IMAP servers, checks whether new e-mails match to your rule sets and then move it to your desired folder (or flags them). It usually doesn't download the full e-mail and _never_ changes the contents of an e-mail.


Demo
----

.. image:: https://asciinema.org/a/294922.svg
    :alt: Basic Demo
    :target: https://asciinema.org/a/294922


Contributing
------------

Bug reports and pull requests are welcome!

In general:

1. Fork this repo on Github
2. Add changes, add test, update docs
3. Submit your pull request (PR) on Github, wait for feedback

But it’s better to `file an issue <https://github.com/arnisoph/postisto/issues/new>`_ with your idea first.

Installing
----------

Download from Github
''''''''''''''''''''

You can download `pre-built binaries <https://github.com/arnisoph/postisto/releases>`_ and `Docker images <https://github.com/arnisoph/postisto/packages>`_ from the Github project page.

Install from Source
'''''''''''''''''''

You want to patch the source code and use your self-built binary? Easy!

.. image:: https://asciinema.org/a/294920.svg
    :alt: Installing from source
    :target: https://asciinema.org/a/294920


Testing
-------

You need a running Docker daemon with Internet connection. The `container image <https://hub.docker.com/r/bechtoldt/tabellarius_tests-docker/>`_ that is beeing downloaded contains Dovecot and Redis.

Start the tests:

.. image:: https://asciinema.org/a/294919.svg
    :alt: make test
    :target: https://asciinema.org/a/294919


Configuring
-----------

Supported Protocols
'''''''''''''''''''

IMAP over Plain Text Transport (don't use it!):

::

    accounts:
      myaccount:
        enable: true
        connection:
            server: imap.server.de
            username: imap@account.de
            password: mypassword
            port: 143
            starttls: false
            imaps: false

IMAP via STARTTLS (usually port 143):

::

    accounts:
      myaccount:
        enable: true
        connection:
            server: imap.server.de
            username: imap@account.de
            password: mypassword
            port: 143
            starttls: true
            imaps: false

IMAP via Force-TLS/SSL (usually port 993):

::

    accounts:
      myaccount:
        enable: true
        connection:
            server: imap.server.de
            username: imap@account.de
            password: mypassword
            port: 993
            starttls: false
            imaps: true

Authentication
''''''''''''''

Plain text in configuration file (don't use it!):

::

    accounts:
      myaccount:
        enable: true
        connection:
            server: imap.server.de
            username: imap@account.de
            password: mypassword
            port: 993
            starttls: false
            imaps: true

Read plain text password from filesystem:

::

    $ ls -l config/
    total 8
    -rw-r--r--  1 ab  staff  15 Jan 20 22:37 config.yml
    $ cat config.yml
    accounts:
      myaccount:
        server: imap.server.de
        username: imap@account.de
        port: 993
        starttls: false
        imaps: true

    $ echo -n "MyP@ssw0rd42" > config/.postisto.myaccount.pwd
    $ ls -lA config
    total 16
    -rw-r--r--  1 ab  staff  12 Jan 20 22:37 .postisto.myaccount.pwd
    -rw-r--r--  1 ab  staff  15 Jan 20 22:37 config.yml
    $ postisto -c config/

The *pwd file* must match ``.postisto.<YOUR-ACCOUNT-NAME-FROM-CONFIG-FILE>.pwd``.


Filters/ Rule Sets
''''''''''''''''''

The `config/ directory <https://github.com/arnisoph/postisto/tree/master/config>`_ in the source code repository contains some useful examples. You can also find more advanced examples in the `tests <https://github.com/arnisoph/postisto/tree/master/test/data/configs/valid>`_.


.. |license| image:: https://img.shields.io/badge/license-Apache--2.0-blue.svg
    :alt: Apache-2.0-licensed
    :target: https://github.com/arnisoph/postisto/blob/master/LICENSE

.. |release| image:: https://img.shields.io/github/v/release/arnisoph/postisto?sort=semver
    :alt: GitHub release (latest SemVer)

.. |gitter| image:: https://badges.gitter.im/arnisoph/postisto.svg
    :alt: Join Gitter Chat
    :target: https://gitter.im/arnisoph/postisto?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge

.. |build| image:: https://img.shields.io/github/workflow/status/arnisoph/postisto/main/master
    :alt: GitHub Workflow Status (branch)

.. |godocs| image:: https://img.shields.io/badge/godoc-reference-blue.svg
    :alt: Go Docs
    :target: https://godoc.org/github.com/arnisoph/postisto

.. |gomod| image:: https://img.shields.io/github/go-mod/go-version/arnisoph/postisto
    :alt: GitHub go.mod Go version

.. |codecov| image:: https://codecov.io/gh/arnisoph/postisto/branch/master/graph/badge.svg
    :alt: codecov badge
    :target: https://codecov.io/gh/arnisoph/postisto

.. |goreport| image:: https://goreportcard.com/badge/github.com/arnisoph/postisto
    :alt: Go Report Card
    :target: https://goreportcard.com/report/github.com/arnisoph/postisto

.. |codebeat| image:: https://codebeat.co/badges/a8d3231c-ee9c-40f5-9bf9-450854a3567a
    :alt: codebeat badge
    :target: https://codebeat.co/projects/github-com-arnisoph-postisto-master