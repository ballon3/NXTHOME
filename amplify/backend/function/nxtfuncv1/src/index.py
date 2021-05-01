import os
from fastapi import FastAPI
from mangum import Mangum
from starlette.middleware.cors import CORSMiddleware
import boto3
from uuid import uuid4

TABLE = os.environ.get("STORAGE_NXTDYNAMO")

client = boto3.client("dynamodb",)

app = FastAPI(title="FastAPI Mangum Example", version='1.0.0')

app.add_middleware(
    CORSMiddleware,
    allow_origins='*',
    allow_credentials=False,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/nxts")
def hello_word():
    return {"Hello": "World"}

handler = Mangum(app)

def handler(event, context):
    event['requestContext'] = {}  # Adds a dummy field; mangum will process this fine
    
    asgi_handler = Mangum(app)
    response = asgi_handler(event, context)

    return response