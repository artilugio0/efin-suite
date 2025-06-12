#!/usr/bin/env python

import argparse
import urllib.request
import http.client

url = r'''https://{{ .Host }}{{ .URL }}'''
method = r'''{{ .Method }}'''
headers = {
{{- range .Headers }}
    r'''{{ .Name }}''': r'''{{ .Value }}''',
{{- end }}
}
body = br'''{{ printf "%s" .Body }}'''

class NoRedirectHandler(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        raise urllib.error.HTTPError(
            req.full_url, code, msg, headers, fp
        )

def make_request(method='GET', url=None, headers=None, body=None):
    opener = urllib.request.build_opener(NoRedirectHandler())
    urllib.request.install_opener(opener)

    result = {
        'status_code': None,
        'reason': None,
        'headers': {},
        'body': ''
    }

    try:
        req = urllib.request.Request(
            url,
            headers=headers or {},
            method=method,
            data=body,
        )
        with urllib.request.urlopen(req) as response:
            result['status_code'] = response.getcode()
            result['reason'] = http.client.responses.get(result['status_code'], 'Unknown')
            result['headers'] = dict(response.getheaders())
            result['body'] = response.read().decode('utf-8', errors='replace')
    except urllib.error.HTTPError as e:
        result['status_code'] = e.code
        result['reason'] = http.client.responses.get(result['status_code'], 'Unknown')
        result['headers'] = dict(e.headers)
        result['body'] = e.read().decode('utf-8', errors='replace')

    return result

def print_response(response_dict):
    status_code = response_dict['status_code']
    reason = response_dict['reason']
    headers = response_dict['headers']
    body = response_dict['body']

    raw_response = f"HTTP/1.1 {status_code} {reason}\r\n"
    for header_name, header_value in headers.items():
        raw_response += f"{header_name}: {header_value}\r\n"
    raw_response += "\r\n"
    raw_response += body

    print(raw_response)

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
        description=f'Make a GET request to {url}',
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

    if args.print_request:
        print_request(args.method, args.url, headers, body)

    response = make_request(args.method, args.url, headers, body)

    if args.print_response:
        print_response(response)
    else:
        print(f'Status: {response["status_code"]} {response["reason"]}')
