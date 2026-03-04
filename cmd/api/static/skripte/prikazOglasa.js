let oordinatePocetak = []
let potrebnoIgracaPocetak = -1


const otkaziDugme = document.querySelector('#otkazi-dugme')
const potvrdiDugme = document.querySelector('#potvrdi-dugme')
const mapaOznaka = document.querySelector('#mapa-oznaka')
const potrebnoIgracaPolje = document.querySelector('#potrebno-igraca-vrednost')
const potrebnoIgracaInput = document.querySelector('#potrebno-igraca-input') 
const povecajBrojIgracaDugme = document.querySelector('#povecaj-broj-igraca')
const smanjiBrojIgracaDugme = document.querySelector('#smanji-broj-igraca')
const dodajKomentarInput = document.querySelector('#komentar-input')

document.addEventListener('DOMContentLoaded', () => {
    koordinatePocetak = mapaOznaka.getAttribute('data-geometry')
    potrebnoIgracaPocetak = +potrebnoIgracaPolje.innerText

    setTimeout(() => {
        const mapa = window.hyperleaflet.map

        mapa.on('click', (e) => {
            const mapaPodaci = document.querySelector('#mapa-podaci')
            const koordinateInput = document.querySelector('#izmeni-oglas-koordinate')
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
            otkaziDugme.disabled = false
            potvrdiDugme.disabled = false
        })
    }, 100)

    otkaziDugme.addEventListener('click', (e) => {
        e.preventDefault()
        mapaOznaka.setAttribute('data-geometry', koordinatePocetak)
        potrebnoIgracaPolje.innerText = potrebnoIgracaPocetak
    })

    povecajBrojIgracaDugme.addEventListener('click', (e) => {
        e.preventDefault()
        potrebnoIgracaInput.value++

        if (+potrebnoIgracaInput.value > +potrebnoIgracaInput.max)
            potrebnoIgracaInput.value = potrebnoIgracaInput.max

        potrebnoIgracaPolje.innerText = potrebnoIgracaInput.value

        otkaziDugme.disabled = false
        potvrdiDugme.disabled = false
    })

    smanjiBrojIgracaDugme.addEventListener('click', (e) => {
        e.preventDefault()
        potrebnoIgracaInput.value--

        if (+potrebnoIgracaInput.value < +potrebnoIgracaInput.min)
            potrebnoIgracaInput.value = potrebnoIgracaInput.min

        potrebnoIgracaPolje.innerText = potrebnoIgracaInput.value

        otkaziDugme.disabled = false
        potvrdiDugme.disabled = false
    })

    dodajKomentarInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey)
            e.preventDefault()
    })
})