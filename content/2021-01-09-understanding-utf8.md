#  Understanding the UTF-8
Many programmers are confused so long about the variation in *UTF-8*, *Unicode*, *ASCII*, *CP936*, *GB2312*, etc.  Why we were always being recommended use UTF-8 for our code in many situations? Character encoding is not a hard problem but it's unclear. To understand the character encoding problem clearly, we should separate the encoding method and character set first. 

Just like Joel says: *Please do not write another line of code until you finish
reading [this article](https://www.joelonsoftware.com/2003/10/08/the-absolute-minimum-every-software-developer-absolutely-positively-must-know-about-unicode-and-character-sets-no-excuses/)*. I recommend you take 15 minutes to read my article, it's easy than Joel's.

In the ancient computer era(1970?), the world was simple and computer was born in US. The computer scientists(or a soft engineer?) treat the world just use English in computer. They designed all of information can be represent enough by 26 alphabets and other symbols; so that they invented ASCII as 127 code points which just include English alphabets, punctuations, and other control symbols.

As we know, computer store everything as Binary, which we call them as a bit 0 or 1. we can use 2<sup>7</sup>= 128 bits to store all ASCII code points. Today's computer use 2<sup>8</sup>=256 bits to represent a byte. Using 8 bits for a byte can be powered by 2 and to be aligned easy.  That's why a byte is not 2<sup>7</sup> or 2<sup>9</sup>.  If we define a byte as
2<sup>16</sup> or 2<sup>32</sup>, it wastes too space (to store 127 ASCII code ^_^).  Therefore, until today, a byte store ASCII code always start with 0, it just an empty bit.

People then find the computer is hard to represent the other language characters soon. We can use our hands draw any language words on papers, but computer can't. We should convert our language to digit first, like ASCII table.  In 1991, Unicode invented, it defined all language characters to digits into a table. For ASCII compatible, the first 127 code points is equal ASCII. As of March 2020, there was a total of 143,859 characters. we can use 32bits Integer to represent all characters. But it doesn't mean every character need 32bits Integer.

For example, if we have a character 'a' which represent 0x61 the same as ASCII. we can store it just in one byte. If we have a character '文' that represent 0xe69687 in Unicode. we should store it in three bytes(e6 96 87).

For reduce the storage, our bytes length should be variable, and given information to tell this word that has how many bytes.

So, Let's check how UTF-8 implementing.

UTF-8 use one byte to represent ASCII(0~127).  
If a code point large than 127, things became different. It separates two parts, the
one part is start with 11 which represent the number of bytes. The other is start with 10, we called it as follow byte.

Here is UTF-8 encode, x means storage data:

``` c
0000 0000-0000 007F | 0xxxxxxx                               //ASCII
0000 0080-0000 07FF | 110xxxxx 10xxxxxx                      //Two bytes
0000 0800-0000 FFFF | 1110xxxx 10xxxxxx 10xxxxxx             //Three bytes
0001 0000-0010 FFFF | 11110xxx 10xxxxxx 10xxxxxx 10xxxxxx    //Four bytes
```

If a character starts with 110, which means it has two bytes. 1110 means three bytes, 11110
means four bytes. UTF-8 support max 4 bytes encoding; excluding the sign bit, it has 2<sup>21</sup> = 2,097,152 code points.

There is a question here, why the follow byte start with 10?

In network transfer, we send information with one by one byte. if a byte starts with 0, we know this is a ASCII byte. if start with 10, means is a follow byte. If we lost a byte, we can drop the other byte fast, which can prevent to occur the half-word problem. In operate system, if we want to remove a word, we can find the byte start without 10, which is easy and simple.  

As you see above, UTF-8 use one-byte oriented to encode every Unicode character. It has two benefits, the one is we don't need to clue the big-endian or little-endian. The other is some old C-library can compatible in UTF-8. `strcmp` can work, because we can compare every word by byte; but `strlen` is not, because many Unicode words didn't just store in one byte.

That's why UTF-8 are spreading the most popular, it has these benefits:  

1. fully comparable ASCII  
2. variant length encoding  
3. error-tolerance, encode and decode easy  
4. byte-oriented, no byte order problems  

You can treat every charset encode method as two parts, the one is a character table, the other is store the code point. Use this method, the term cp936, gb2312, and other encode methods will not confuse you.
