:Date: 10/26/2015

WSGI(pep 3333) 笔记
===================

-  | the goal of WSGI is to fgacilitate easy interconnection of existing
     servers
   | and applications or frameworks, not to create a new web framework.

-  | the WSGI has two sides: the "server" or "gateway" side, and the
     "application"
   | or "framework" side.

-  | a server, gateway, or application that is invoking a callable *must
     not* have
   | any dependecy on what kind of callable was provided to it.
     callables are only
   | to be called, not introspected upon.

-  | when you see the word "string" in the document, it refers to a
     "native" string
   | (in py2k, it's bytes, in py3k, it's unicode);
   | when you see references to "bytestring", it references to
     "bytestring"(in py2k,
   | it's str, in py3k, it's bytes).

-  | Application objects must be able to be invoked more than once, as
     virtually
   | all servers/gateways (other than CGI) will make such repeated
     request.

-  | a server or gateway *must* invoke the application object using
     positional(not keyword)
   | arguments. e.g. ``result = application(environ, start_response)``.

-  | the application *must* incoke the ``start_response`` callable using
     positional arguments.
   | e.g. ``start_response(status, response_headers)``

-  | the ``start_response`` callable must return a ``write(body_data)``
     callable that
   | takes one positional parameterL a bytestring to be written as part
     of the HTTP
   | response body.

-  | when called by the server, the application object must return an
     iterable
   | yielding zero or more bytestrings.

-  | if the iterable returned by the application has a close() method,
     the server
   | or gateway musit call that method upon completion of the current
     request.

-  | the application *must* incoke the ``start_response()`` callable
     before the
   | iterable yields its first body bytestring, so that the server can
     send the
   | headers before any body content; but the server *must not* assume
     that
   | ``start_response()`` has been called before they begin iterating
     over the
   | iterable.

-  | in general, the server or gateway is responsible for ensuring that
     corrent
   | headers are sent to the client.

-  | all encoding/decoding must be handled by the application, all
     string passed
   | to or from the server must be of type str or bytes, nerver unicode.
