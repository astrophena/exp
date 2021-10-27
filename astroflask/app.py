from flask import Flask, render_template, request, make_response
import os

app = Flask(__name__)

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
    return render_template('sysinfo.html', uname=f'{os.uname()}')
