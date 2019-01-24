package bir

var loginEnvelope = `
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:ns="http://CIS/BIR/PUBL/2014/07">     <soap:Header xmlns:wsa="http://www.w3.org/2005/08/addressing">                                                  <wsa:To>https://wyszukiwarkaregontest.stat.gov.pl/wsBIR/UslugaBIRzewnPubl.svc</wsa:To>                          <wsa:Action>http://CIS/BIR/PUBL/2014/07/IUslugaBIRzewnPubl/Zaloguj</wsa:Action>                                 </soap:Header>                                                                                                     <soap:Body>                                                                                                              <ns:Zaloguj>                                                                                                         <ns:pKluczUzytkownika>{{ .APIKey }}</ns:pKluczUzytkownika>                                                 </ns:Zaloguj>                                                                                           </soap:Body>                                                                                                 </soap:Envelope>
`

type login struct {
	APIKey string
}

var searchEnvelope = `
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:ns="http://CIS/BIR/PUBL/2014/07" xmlns:dat="http://CIS/BIR/PUBL/2014/07/DataContract">
<soap:Header xmlns:wsa="http://www.w3.org/2005/08/addressing">
<wsa:To>https://wyszukiwarkaregontest.stat.gov.pl/wsBIR/UslugaBIRzewnPubl.svc</wsa:To>
<wsa:Action>http://CIS/BIR/PUBL/2014/07/IUslugaBIRzewnPubl/DaneSzukaj</wsa:Action>
</soap:Header>
   <soap:Body>
      <ns:DaneSzukaj>
         <ns:pParametryWyszukiwania>
            <dat:Nip>{{ .Nip }}</dat:Nip> 
         </ns:pParametryWyszukiwania>
      </ns:DaneSzukaj>
   </soap:Body>
</soap:Envelope>   
`

type search struct {
	Nip string
}
