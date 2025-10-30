# JSxcite
*extract things from js*

`jsxcite` heavily inspired by [jsluice](https://github.com/BishopFox/jsluice) and therefore the purpose of the tool is to
make life easier when analyzing JavaScript files for hidden paths and secrets.

### Warning - the tool is under development

## But why?
Despite all the charm of `jsluice`, in my practice it missed a significant number of potential paths. 
Therefore, `jsxcite` aims to keep misses to a minimum.
The difference in the comparison of the results of the utilities is already noticeable but there is still work to do.

## Installation

`go install github.com/indigo-sadland/jsxcite@latest`

## Usage
- **Paths** mode (single file)\
    `jsxcite paths -t https://example.com/libs/main.js `\
    `jsxcite paths -t /local/dir/main.js`
- **Paths** mode  (multiple files)\
    `find /local/dir/ -type f -name "*.js" | jsxcite paths`


You can pipe the output with `jq` for better formatting\
`find /local/dir/ -type f -name "*.js" | jsxcite paths | jq`