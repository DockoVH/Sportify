const potrebnoIgracaInput = document.querySelector('#novi-oglas-potrebno-igraca')

potrebnoIgracaInput.addEventListener('change', () => {
    potrebnoIgracaInput.value = Math.floor(potrebnoIgracaInput.value)

    if (+potrebnoIgracaInput.value < +potrebnoIgracaInput.min)
        potrebnoIgracaInput.value = potrebnoIgracaInput.min

    if (+potrebnoIgracaInput.value > +potrebnoIgracaInput.max)
        potrebnoIgracaInput.value = potrebnoIgracaInput.max
})

document.addEventListener('DOMContentLoaded', () => {
    setTimeout(() => {
        const mapa = window.hyperleaflet.map

        mapa.on('click', (e) => {
            const mapaPodaci = document.querySelector('#mapa-podaci')
            const koordinateInput = document.querySelector('#novi-oglas-koordinate')
            const koordinate = [e.latlng.lat, e.latlng.lng]
            
            if (mapaPodaci.childNodes.length === 1)
            {
                const point = document.createElement('span')
                point.setAttribute('data-id', '1')
                point.setAttribute('data-geometry-type', 'Point')
                point.setAttribute('data-geometry', `[${koordinate}]`)

                mapaPodaci.prepend(point)
            }
            else
            {
                const point = mapaPodaci.querySelector('span')
                point.setAttribute('data-geometry', `[${koordinate}]`)
            }

            koordinateInput.value = koordinate
        })
    }, 100)
})