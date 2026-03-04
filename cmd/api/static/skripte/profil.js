const slikaContainer = document.querySelector('#nova-slika')
const fileInput = slikaContainer.querySelector('input[type=file]')
const textInput = slikaContainer.querySelector('input[type=text]')
const slika = slikaContainer.querySelector('img')

const encodeImageFileAsURL = (element) => {
    const file = element.files[0]
    const reader = new FileReader()
    reader.addEventListener('loadend', () => {
        slika.src = reader.result
        textInput.value = reader.result
    })

    reader.readAsDataURL(file)
}

fileInput.addEventListener('change', () => encodeImageFileAsURL(fileInput))

slikaContainer.addEventListener('click', () => fileInput.click())