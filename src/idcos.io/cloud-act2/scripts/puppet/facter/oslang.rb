
Facter.add(:oslang) do
  confine :kernel => 'Linux'
  setcode do
    encoding = 'utf-8'
    if File.exist? '/etc/locale.conf'
          result = Facter::Util::Resolution.exec('cat /etc/locale.conf')
          encoding = result.split('=')[1].downcase if result.split('=')[1]
    else
         line = Facter::Util::Resolution.exec('locale')
         line.split('\n').each do |f|
             if (match = f.match(/^(LANG)\=(.*)/))
                 encoding = match[2].downcase if match[2]
             end
         end
    end
    encoding
  end
end

Facter.add(:oslang) do
  confine :kernel => 'windows'
  setcode do
      require 'win32/registry'
      encoding = "cp936"
      reg = Win32::Registry::HKEY_LOCAL_MACHINE.open('SYSTEM\CurrentControlSet\Control\Nls\CodePage')
      H = Hash["936"=>"cp936","950"=>"cp950","65000"=>"utf-7","65001"=>"utf-8",\
        "54936"=>"GB18030","52936"=>"hz-gb-2312","12000"=>"utf-32","1200"=>"utf-16"]
      result = reg['OEMCP'] if reg['OEMCP']
      reg.close
      encoding = H[result].downcase if H[result]
      encoding
  end
end

Facter.add(:oslang) do
  confine :kernel => 'AIX'
  setcode do
      #iconv  -l
       encoding ='utf-8'
       line = Facter::Util::Resolution.exec('locale')
       line.split('\n').each do |f|
           if (match = f.match(/^(LANG)\=(.*)/))
               encoding = match[2].downcase if match[2]
           end
       end
       encoding
   end
end
