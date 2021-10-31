from flask import Flask, render_template, request, make_response
import os

app = Flask(__name__)

@app.errorhandler(404)
def not_found(e):
    return render_template('not_found.html')

@app.route('/', methods=['GET'])
def index():
    return render_template('index.html')

@app.route('/greet', methods=['POST'])
def greet():
    return render_template('greet.html', name=request.form.get('name', 'world'))

@app.route('/ua', methods=['GET'])
def ua():
    response = make_response(request.headers.get('User-Agent'), 200)
    response.mimetype = 'text/plain'
    return response

@app.route('/sysinfo', methods=['GET'])
def sysinfo():
    uname = os.uname()
    if request.args.get('format') == 'json':
        return {
            # https://docs.python.org/3/library/os.html#os.uname
            'sysname': uname.sysname,
            'nodename': uname.nodename,
            'release': uname.release,
            'version': uname.version,
            'machine': uname.machine,
        }
    return render_template('sysinfo.html', uname=uname)
