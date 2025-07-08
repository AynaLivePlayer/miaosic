# Miaosic

Music Provider Repository, provide a universal interface for different music service.

## Command line Tool

please check [miaosic cmd tool](./cmd/miaosic/README.md)

## How to Use

please figure it out by yourself.

## Available Providers

| Provider       | Search | Info | Url | Lyric | Playlist | loginable |
|----------------|--------|------|-----|-------|----------|-----------|
| netease        | ✓    * | ✓  * | ✓ * | ✓ *   | ✓ *      | ✓         |
| kuwo           | ✓      | ✓    | ✓   | ✓     | ✗        | ✗         |
| kugou          | ✓      | ✓    | ✓ * | ✓     | ✓        | ✓         |
| bilibili-video | ✓      | ✓    | ✓   | ✓     | ✓        | ✗         |
| qq             | ✓*     | ✓*   | ✓*  | ✓*    | ✓*       | ✓         |

> \* means require login

## Known Problem

1. Current implementation of source registration is **not** threading-safe, 
please implement thread-safe version by yourself in your project require multi-thread access (for example, web).

## Disclaimer

All APIs used in this project are  **publicly available** on the internet and not obtained through illegal means such as
reverse engineering.

The use of this project may involve access to copyrighted content. This project does **not** own or claim any rights to
such content. **To avoid potential infringement**, all users are **required to delete any copyrighted data obtained
through this project within 24 hours.**

Any direct, indirect, special, incidental, or consequential damages (including but not limited to loss of goodwill, work
stoppage, computer failure or malfunction, or any and all other commercial damages or losses) that arise from the use or
inability to use this project are **solely the responsibility of the user**.

This project is completely free and open-source, published on GitHub for global users for **technical learning and
research purposes only**. This project does **not** guarantee compliance with local laws or regulations in all
jurisdictions.

**Using this project in violation of local laws is strictly prohibited.** Any legal consequences arising from
intentional or unintentional violations are the user's responsibility. The project maintainers accept **no liability**
for such outcomes.
