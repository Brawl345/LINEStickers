#!/usr/bin/env python3
import os
from argparse import ArgumentParser

from requests import get

META_URL = "https://sdl-stickershop.line.naver.jp/stickershop/v1/product/{0}/android/productInfo.meta"
ANIMATED_URL = "https://sdl-stickershop.line.naver.jp/products/0/0/1/{0}/android/animation/{1}.png"


class Stickerpack:
    """Defines information about a LINE sticker/stamp pack"""

    def __init__(self, json):
        """
        Initializes the Image class with information for a sticker/stamp pack
        :param json: String. Raw JSON
        """
        if "packageId" not in json:
            raise Exception("This is not a LINE sticker pack")

        self.packageId = json["packageId"]
        if "en" in json["title"]:
            self.title = json["title"]["en"]
        else:
            self.title = json["title"]["ja"]

        if "en" in json["author"]:
            self.author = json["author"]["en"]
        else:
            self.author = json["author"]["ja"]

        self.animated = False
        if "hasAnimation" in json:
            self.animated = json["hasAnimation"]

        self.sound = False
        if "sound" in json:
            self.sound = json["hasSound"]

        self.stickers = []
        self.amount = 0
        for sticker in json["stickers"]:
            self.stickers.append(Sticker(sticker))
            self.amount += 1


class Sticker:
    """Defines information about a LINE sticker/stamp"""
    STICKER_URL = "https://stickershop.line-scdn.net/stickershop/v1/sticker/{0}/android/sticker.png;compress=true"

    def __init__(self, json):
        """
        Initializes the Image class with information for a sticker/stamp
        :param json: String. Raw JSON
        """
        if "id" not in json:
            raise Exception("This is not a LINE sticker")

        self.id = json["id"]
        self.width = json["width"]
        self.height = json["height"]
        self.download_url = self.STICKER_URL.format(self.id)


def main(storeid, static):
    if not isinstance(storeid, int):
        print("This is not an integer")
        return

    print("Getting LINE sticker pack...")
    req = get(META_URL.format(storeid))
    if req.status_code != 200:
        print("=> Received HTTP error {0}".format(req.status_code))
        return

    stickerpack = Stickerpack(req.json())
    folderpath = os.path.join("output", "LINE_{0}".format(stickerpack.packageId))
    print("=> {0} by {1}".format(stickerpack.title, stickerpack.author))

    if not os.path.exists(folderpath):
        try:
            os.makedirs(folderpath)
        except OSError as e:
            print("Error while creating directory: {0}".format(e))
            return

    for i, sticker in enumerate(stickerpack.stickers):
        print("  Downloading {0}/{1}...".format(i + 1, stickerpack.amount))
        if stickerpack.animated and not static:
            sticker.download_url = ANIMATED_URL.format(stickerpack.packageId, sticker.id)
        sticker_download = get(sticker.download_url)
        if sticker_download.status_code != 200:
            print("    => FAILED!")
        else:
            with open(os.path.join(folderpath, "{0}.png".format(sticker.id)), "wb") as stickerfile:
                stickerfile.write(sticker_download.content)
                print("    => SUCCESS!")

    print("Writing info file...")
    with open(os.path.join(folderpath, "info.txt"), "w+") as infofile:
        infofile.write("{0} by {1}\n".format(stickerpack.title, stickerpack.author))
        infofile.write("Meta URL: {0}\n\n".format(META_URL.format(storeid)))
        infofile.write("{0} stickers:\n".format(stickerpack.amount))
        for sticker in stickerpack.stickers:
            infofile.write(sticker.download_url + "\n")
    with open(os.path.join(folderpath, "info.json"), "wb") as infojson:
        infojson.write(req.content)
    print("DONE!")


if __name__ == "__main__":
    parser = ArgumentParser()
    parser.add_argument('storeid', type=int, help="LINE Store ID")
    parser.add_argument(
            '--static',
            action='store_true',
            default=False,
            dest='static',
            help='Always download static PNGs'
    )
    arguments = parser.parse_args()
    main(arguments.storeid, arguments.static)
