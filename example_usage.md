mutation createPolishWord{
	createPolishWord(word: "jedzenie"){
    id
    text
  }  
}
query getPolishWords{
  polishWords{
    id
    text
  }
}
mutation createEnglishWord{
	createEnglishWord(word: "bamboo"){
    id
    text
  }  
}
query getEnglishWords{
  englishWords{
    id
    text
  }
}
mutation createTranslation{
  createTranslation(translation: {
    polishWord: "wieża",
    englishWord: "rook",
    examples: [
      {
        text: "If rook didn't move castle is enabled."
        inPolish: false
      },
      {
        text: "Rook is the stronges piece after queen.",
        inPolish: false
      }
    ]
    })
  {
    polishWord {
      id
      text
    }
    englishWord {
      id
      text
    }
    examples {
      id
      text
      inPolish
    }
  }
}
query getTranslations {
  translations{
    id
    polishWord{
      text
    }
    englishWord {
      text
    }
    examples {
      id
      text
      inPolish
    }
  }
}
mutation createExample{
  createExample(example:{
    translationID: 1,
    example: {
      text: "W wieży często spotkać można maga.",
      inPolish: true
    }
  })
  {
    id
    inPolish
    text
  }
}
query getTranslationForPolishWord {
  translationToEnglish(wordInPolish: "wieża")
  {
    id
    polishWord{
      text
    }
    englishWord {
      text
    }
    examples {
      text
      inPolish
    }
  }
}
query getTranslationForEnglishWord {
  translationToPolish(wordInEnglish: "rook")
  {
    id
    polishWord{
      text
    }
    englishWord {
      text
    }
    examples {
      text
      inPolish
    }
  }
}
mutation deletePolishWord {
  deletePolishWord(id: 7)
}
mutation deleteEnglishWord {
  deleteEnglishWord(id: 7)
}
mutation deleteTranslation {
  deleteTranslation(id: 1)
}
mutation deleteExample{
  deleteExample(id: 11)
}
mutation changePolishText{
  updatePolishWordText(id: 1, text: "testy")
  {
    id
    text
  }
}
mutation changeExampleText{
  updateExampleText(id: 8, text: "Rook is worth 5 points.")
  {
    id
    text
  }
}
mutation changeEnglishText{
  updateEnglishWordText(id: 6, text: "tower")
  {
    id
    text
  }
}
query getTranslationById{
  getTranslation(id: 2)
  {
    id
    englishWord{id, text}
    polishWord{id, text}
    examples{id, text, inPolish}
  }
}
query getExampleById{
  getExample(id: 50)
  {
    id
    text
    inPolish
    translationID
  }
}
query getPolishWordById{
  getPolishWord(id: 1){
  	id
    text
  }
}
query getEnglishWordById{
  getEnglishWord(id: 1){
  	id
    text
  }
}