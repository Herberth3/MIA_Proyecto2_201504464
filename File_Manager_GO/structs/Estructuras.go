package structs

type Partition struct {
	Part_status [1]byte
	Part_type   [1]byte
	Part_fit    [1]byte
	Part_start  [35]byte
	Part_size   [35]byte
	Part_name   [26]byte
}

type MBR struct {
	Mbr_size           [35]byte
	Mbr_date_created   [20]byte
	Mbr_disk_signature [35]byte
	Mbr_disk_fit       [1]byte
	Mbr_partition      [4]Partition
}

type EBR struct {
	Part_status [1]byte
	Part_fit    [1]byte
	Part_start  [35]byte
	Part_size   [35]byte
	Part_next   [35]byte
	Part_name   [26]byte
}

// STRUCTS PARA EL SISTEMA DE ARCHIVOS

type SuperBloque struct {
	S_filesystem_type   [1]byte  //Guarda el numero que identifica al sistea de archivos utilizados
	S_inodes_count      [35]byte //Guarda el número total de inodos
	S_blocks_count      [35]byte //Guarda el número total de bloques
	S_free_blocks_count [35]byte //Contiene el número de bloques libres
	S_free_inodes_count [35]byte //Contiene el número de inodos libres

	S_mtime  [25]byte //Última fecha en el que el sistema fue montado
	S_umtime [25]byte //Última fecha en que el sistema fue desmontado

	S_mnt_count      [6]byte  //Indica cuantas veces se ha montado el sistema
	S_magic          [10]byte //Valor que identifica, tendrá el valor 0xEF53
	S_inode_size     [35]byte //Tamaño del inodo
	S_block_size     [35]byte //Tamaño del bloque
	S_first_ino      [35]byte //Primer inodo libre
	S_first_blo      [35]byte //Primer bloque libre
	S_bm_inode_start [35]byte //Guardará el inicio del bitmap de inodos
	S_bm_block_start [35]byte //Guardará el inicio del bitmap de bloques
	S_inode_start    [35]byte //Guardará el inicio de la tabla de inodos
	S_block_start    [35]byte //Guardará el inicio de la tabla de bloques
}

type InodoTable struct {
	I_uid  [7]byte  //UID del usuario propietario del archivo/carpeta
	I_gid  [7]byte  //GID del grupo al que pertenece el archivo/carpeta
	I_size [35]byte //Tamaño del archivo en bytes

	I_block [16]byte //Array de bloques
	I_type  [1]byte  //Indica si es archivo o carpeta 1=archivo, 0=carpeta
	I_perm  [3]byte  //Guardara los permisos del archivo o carpeta

	I_atime [25]byte //Ultima fecha de lectura del inodo sin modificarlo
	I_ctime [25]byte //Fecha en la que se creo el inodo
	I_mtime [25]byte //Ultima decha de modificacion

}

type Content struct {
	B_name  [12]byte //Nombre de la carpeta o archivo
	B_inodo [4]byte  //Apuntador hacia un inodo asociado al archivo o carpeta
}

type BloqueCarpeta struct {
	B_content [4]Content //Array con el contenido de la carpeta
}

type BloqueArchivo struct {
	B_content [64]byte //Array con el contenido del archivo
}

type BloqueApuntadores struct {
	B_pointer [16]byte //Array con los apuntadores hacia bloques
}
