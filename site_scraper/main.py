import asyncio
import time
import aiohttp
from bs4 import BeautifulSoup
from urllib.parse import urlparse
import logging
from random import randint
import logging
import socket
import os
import websockets
import json
from collections import namedtuple
from aiocsv import AsyncWriter
import aiofiles


CSV_FILE = os.environ.get('SHARE_FOLDER') or ''
CSV_FILE += "data.txt"

logging.basicConfig(format='%(asctime)s %(levelname)s:%(message)s',  datefmt='%m/%d/%Y %I:%M:%S %p', level=logging.DEBUG)

# range for random timeout between requests
timeout = (1, 3)

# start urls for scrapping
urls = ["https://rosexperts.ru/"]
urls_list_lock = asyncio.Lock()

# exclude urls containing these words
skipping_parh = ["@", "/uploads/"]

# hashes for unique content
hashes = []
hashes_list_lock = asyncio.Lock()

Func = namedtuple('Func', 'AppendText')
func = Func('AppendText')


async def main():

    ws = None
    # async with aiofiles.open(CSV_FILE, 'w'):
    #     pass
    socket_host = os.environ['SOCKET_HOST']
    socket_port = int(os.environ['SOCKET_PORT'])

    while True:
        try:
            ws = await websockets.connect(f"ws://{socket_host}:{socket_port}/")
        except Exception as err:
            logging.info(f"Failed connect to socket. {err}. Keep trying...")
            await asyncio.sleep(5)
        else:
            logging.info(f"connected to socket.")
            break

    logging.info("starting scraping...")
    start_time = time.time()
    tasks = []

    for url in urls:
        domen = urlparse(url).netloc
        task = asyncio.create_task(scrape(url, domen, ws))
        tasks.append(task)

    await asyncio.gather(*tasks)

    logging.info(f"scraping time {time.time()-start_time:.2f}")
    logging.info(f"closing socket.")
    # await ws.close(1000)


async def exec(func: str, arg: str, ws) -> str:
    response = ""
    try:
        await ws.send(json.dumps({"Func": func, "Args": arg}))
        response = await ws.recv()
    except Exception as err:
        response = str(err)
    return response


async def scrape(url:str, domen:str, ws):

    headers = {
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) "
            "Chrome/113.0.0.0 Safari/537.36 Edg/113.0.1774.35"
   }

    # wait a little to not scare the server
    await asyncio.sleep(randint(*timeout))

    # make request to the server
    async with aiohttp.ClientSession(headers=headers) as session:
        async with session.get(url) as response:
            if response.status == 200:

                logging.info(f"Parsing url {url}")

                # get response text
                body = await response.text()
                soup = BeautifulSoup(body, 'html.parser')
                # get rid of multiple \n
                text = " ".join(soup.find("div", {"class": "content_wrap"}).text.replace("\n", " ").split())
                # get hash of text
                text_hash = hash(text)

                # check if the text is unique
                async with hashes_list_lock:
                    if text_hash in hashes:
                        return
                    else:
                        hashes.append(text_hash)

                res = await exec(func.AppendText, soup.title.string + "/n" + text + "/n", ws)
                if res == "500":
                    logging.info(f"Failed write to database {soup.title.string} ({len(text)} bytes)")
                elif res == "200":
                    logging.info(f"{soup.title.string} with length {len(text)} has wroted")
                else:
                    logging.info(f"Something else happened")

                # async with aiofiles.open(CSV_FILE, 'a+') as f:
                #     await f.write(text)
                #     await f.write("\n\n")
                #     # writer = AsyncWriter(f, delimiter=";")  # ,
                #     # await writer.writerow([soup.title.string, text+"\n\n"])

                # get all links from the page
                page_urls = soup.find_all('a', href=True)
                # get unique links
                page_urls = list(set(page_urls))
                # get links as lowcased strings
                page_urls = [page_url.get('href').lower() for page_url in page_urls]
                # filter external links and links from excluding list
                page_urls = [page_url for page_url in page_urls if domen in page_url and not [skip for skip in skipping_parh if skip in page_url]]
                # filter external links and links from excluding list
                new_urls = []
                # exclude processed links
                async with urls_list_lock:
                    new_urls = [new_url for new_url in page_urls if new_url not in urls]
                    # and write remaining links to the list of processed links
                    urls.extend(new_urls)

                # too fast for server
                # await asyncio.gather(*[asyncio.create_task(scrape(new_url, domen)) for new_url in new_urls])

                # carefully start tasks for processing urls
                [await asyncio.create_task(scrape(new_url, domen, ws)) for new_url in new_urls]


loop = asyncio.get_event_loop()
loop.run_until_complete(main())
