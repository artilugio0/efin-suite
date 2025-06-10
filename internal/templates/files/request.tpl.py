#!/usr/bin/env python

import argparse
import urllib.request

url = r'''https://{{ .Host }}{{ .URL }}'''
method = r'''{{ .Method }}'''
headers = {
{{- range .Headers }}
    r'''{{ .Name }}''': r'''{{ .Value }}''',
{{- end }}
}
body = br'''{{ printf "%s" .Body }}'''

def make_request(method=method, url=url, headers=headers, body=body):
    req = urllib.request.Request(
        url,
        headers=headers,
        method=method,
        data=body,
    )

    return urllib.request.urlopen(req)

def print_response(response):
    print(f'HTTP/1.1 {response.status} {response.reason}', end="\r\n")

    for header, value in response.getheaders():
        print(f'{header}: {value}', end="\r\n")

    print(end="\r\n")

    raw_body = response.read()
    print(raw_body.decode('latin1', errors='replace'))

def print_request(method=method, url=url, headers=headers, body=body):
    print(f'{method.upper()} {url} HTTP/1.1', end='\r\n')

    for header, value in (headers or {}).items():
        if header.lower() != 'host':
            print(f'{header}: {value}', end='\r\n')

    print(end='\r\n')

    if body:
        if isinstance(body, bytes):
            print(body.decode('latin1', errors='replace'))
        else:
            print(body)




if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        prog='make_request.py',
        description=f'Make a {{ .Method }} request to {url}',
        epilog='Script generated with Efin: https://github.com/artilugio0/efin-vibes')

    parser.add_argument('-m', '--method', default=method, help='change the method of the request')
    parser.add_argument('-u', '--url', default=url, help='change the url of the request')
    parser.add_argument('-H', '--header', default=[], action='append', help='add a header to the request. Format: "name: value"')
    parser.add_argument('-r', '--remove-header', default=[], action='append', help='remove the specified header')
    parser.add_argument('-b', '--body', type=lambda b: bytes(b, 'utf-8'), help='replace body')
    parser.add_argument('-q', '--print-request', action='store_true', default=False, help='print raw request')
    parser.add_argument('-p', '--print-response', action='store_true', default=False, help='print raw response')

    args = parser.parse_args()

    extra_headers = {h.split(':')[0]: ' '.join(h.split(' ')[1:]) for h in args.header}
    headers = {**headers, **extra_headers}

    remove_headers = [h.lower() for h in args.remove_header]
    headers = {n:v for n, v in headers.items() if n.lower() not in remove_headers}

    if args.body is not None:
        body = args.body

        prev_len = len(headers)
        headers = {n:v for n, v in headers.items() if n.lower() != 'content-length'}
        if len(headers) < prev_len:
            headers['Content-Length'] = str(len(args.body))

    ## Make the request
    if args.print_request:
        print_request(args.method, args.url, headers, body)

    with make_request(args.method, args.url, headers, body) as response:
        if args.print_response:
            print_response(response)
        else:
            print(f'Status: {response.status} {response.reason}')
